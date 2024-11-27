package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

var (
	mixinKeyEncTab = []int{
		46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
		33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
		61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
		36, 20, 34, 44, 52,
	}
	cache          sync.Map
	lastUpdateTime time.Time
)

func FetchVideos(keyword, order string) string {
	urlStr := "https://api.bilibili.com/x/web-interface/wbi/search/type?search_type=video&keyword=" + keyword + "&order=" + order
	newUrlStr, err := signAndGenerateURL(urlStr)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return "Error: " + err.Error()
	}
	req, err := http.NewRequest("GET", newUrlStr, nil)
	fmt.Println(newUrlStr)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return "Error: " + err.Error()
	}
	req.Header.Set("Accept", "*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")

	originCookie := "buvid_fp_plain=undefined; buvid4=18A68FB2-F597-AF65-5F3B-7CF1A758861504264-023041811-ex5UQvvtG351fkucjsxdxw%3D%3D; enable_web_push=DISABLE; CURRENT_BLACKGAP=0; FEED_LIVE_VERSION=V_WATCHLATER_PIP_WINDOW3; CURRENT_QUALITY=80; buvid3=014F6001-B296-6944-8130-3933A874E7EE85205infoc; b_nut=1713335685; _uuid=B49410F89-710F4-8558-752F-65410C48B28C211180infoc; header_theme_version=CLOSE; hit-dyn-v2=1; rpdid=|(k|J|~|muRk0J'u~uYJY|l)k; fingerprint=49495033dc74747a84ddfa9d12d82580; buvid_fp=49495033dc74747a84ddfa9d12d82580; DedeUserID=312811426; DedeUserID__ckMd5=3fdfdd712bd29342; LIVE_BUVID=AUTO7917235357201806; PVID=2; CURRENT_FNVAL=4048; home_feed_column=5; bili_ticket=eyJhbGciOiJIUzI1NiIsImtpZCI6InMwMyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzI3NzY3MjMsImlhdCI6MTczMjUxNzQ2MywicGx0IjotMX0.FfzyEzxclBJzOolzGYx959rQ5Shdf8fbaqNdZrm-zZ8; bili_ticket_expires=1732776663; bili_jct=b475e908600a6f289dfb576a58b2d983; bp_t_offset_312811426=1003644401496358912; b_lsid=7AFB65AC_19366585862; match_float_version=ENABLE; browser_resolution=2048-983"
	cookieVal, ok := getCookieCached()
	if ok {
		originCookie = cookieVal
	}

	req.Header.Set("Cookie", originCookie)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %s", err)
		return "Error: " + err.Error()
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %s", err)
		return "Error: " + err.Error()
	}

	return string(body)
}

func signAndGenerateURL(urlStr string) (string, error) {
	urlObj, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	imgKey, subKey := getWbiKeysCached()
	query := urlObj.Query()
	params := map[string]string{}
	for k, v := range query {
		params[k] = v[0]
	}
	newParams := encWbi(params, imgKey, subKey)
	for k, v := range newParams {
		query.Set(k, v)
	}
	urlObj.RawQuery = query.Encode()
	newUrlStr := urlObj.String()
	return newUrlStr, nil
}

func encWbi(params map[string]string, imgKey, subKey string) map[string]string {
	mixinKey := getMixinKey(imgKey + subKey)
	currTime := strconv.FormatInt(time.Now().Unix(), 10)
	params["wts"] = currTime

	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Remove unwanted characters
	for k, v := range params {
		v = sanitizeString(v)
		params[k] = v
	}

	// Build URL parameters
	query := url.Values{}
	for _, k := range keys {
		query.Set(k, params[k])
	}
	queryStr := query.Encode()

	// Calculate w_rid
	hash := md5.Sum([]byte(queryStr + mixinKey))
	params["w_rid"] = hex.EncodeToString(hash[:])
	return params
}

func getMixinKey(orig string) string {
	var str strings.Builder
	for _, v := range mixinKeyEncTab {
		if v < len(orig) {
			str.WriteByte(orig[v])
		}
	}
	return str.String()[:32]
}

func sanitizeString(s string) string {
	unwantedChars := []string{"!", "'", "(", ")", "*"}
	for _, char := range unwantedChars {
		s = strings.ReplaceAll(s, char, "")
	}
	return s
}

func updateCache() {
	if time.Since(lastUpdateTime).Minutes() < 10 {
		return
	}
	imgKey, subKey := getWbiKeys()
	cache.Store("imgKey", imgKey)
	cache.Store("subKey", subKey)
	lastUpdateTime = time.Now()
}

func getWbiKeysCached() (string, string) {
	updateCache()
	imgKeyI, _ := cache.Load("imgKey")
	subKeyI, _ := cache.Load("subKey")
	return imgKeyI.(string), subKeyI.(string)
}

func UpdateCookie(cookieVal string) {
	fmt.Printf("update cookie: %s", cookieVal)
	cache.Store("cookieKey", cookieVal)
}

func getCookieCached() (string, bool) {
	cookieVal, ok := cache.Load("cookieKey")
	return cookieVal.(string), ok

}

func getWbiKeys() (string, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		fmt.Printf("Error creating request: %s", err)
		return "", ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %s", err)
		return "", ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s", err)
		return "", ""
	}
	json := string(body)
	imgURL := gjson.Get(json, "data.wbi_img.img_url").String()
	subURL := gjson.Get(json, "data.wbi_img.sub_url").String()
	imgKey := strings.Split(strings.Split(imgURL, "/")[len(strings.Split(imgURL, "/"))-1], ".")[0]
	subKey := strings.Split(strings.Split(subURL, "/")[len(strings.Split(subURL, "/"))-1], ".")[0]
	// imgKey := "7cd084941338484aae1ad9425b84077c"
	// subKey := "4932caff0ff746eab6f01bf08b70ac45"
	return imgKey, subKey
}
