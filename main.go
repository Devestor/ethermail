package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

func main() {
	tokenMap := make(map[string]string)
	tokenMap["xxxxxxxxx@ethermail.io"] = "xxxxxxx"

	// Getting a slice
	mailList := []string{}
	for k := range tokenMap {
		mailList = append(mailList, k)
	}

	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatal("Unfortunately can't load a location")
	}

	scheduler := gocron.NewScheduler(location)
	scheduler.Every(1).Day().At("11:59").Do(func() {

		fmt.Println("Starting Send mails")
		SendMail(mailList, tokenMap)
		fmt.Println("End Send mails")

		fmt.Println("Starting Read messages")
		ReadAllMessage(mailList, tokenMap)
		fmt.Println("End Read messages")
	})

	scheduler.StartAsync()
	scheduler.StartBlocking()

}

func ReadAllMessage(mailList []string, tokenMap map[string]string) {
	empData := [][]string{
		{"No.", "Mail", "Mail IDs", "Status", "Updated", "Remark"},
	}

	for k, mail := range mailList {
		token := tokenMap[mail]
		mailboxes := GetMailBoxes(token)
		var mailboxID, mailIDs, remark string
		var resulrMakeRead *ReponseMakeReadAll
		if len(mailboxes.Results) > 0 {
			mailboxID = mailboxes.Results[0].ID
			mailIDs = GetAllMailBox(mailboxID, token)
			resulrMakeRead = MakeReadAll(mailboxID, mailIDs, token)

			remark = "Found"
		} else {
			remark = fmt.Sprintf("%d) Not found mailbox, Mail: %s", k+1, mail)
		}
		log.Printf("%d) GetAllMailBox: Mail: %s, mailIDs: %s, Success: %s, Updated: %s, Remark: %s", k+1, mail, mailIDs, strconv.FormatBool(resulrMakeRead.Success), strconv.Itoa(resulrMakeRead.Updated), remark)

		empData = append(empData, []string{strconv.FormatInt(int64(k+1), 10), mail, mailIDs, strconv.FormatBool(resulrMakeRead.Success), strconv.Itoa(resulrMakeRead.Updated), remark})
	}

	now := time.Now().Format("02-01-2006 15:04:05")
	fileName := "ReadAllMessage_result_" + now + ".csv"
	csvFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	for _, empRow := range empData {
		_ = csvwriter.Write(empRow)
	}
	csvwriter.Flush()
	csvFile.Close()
}

func SendMail(mailList []string, tokenMap map[string]string) {
	empData := [][]string{
		{"No.", "From Address", "To Address", "Response"},
	}

	// Sending mails
	for _, mail := range mailList {
		fromAddress := mail

		// Random mails for send
		rand.NewSource(time.Now().UnixNano())
		numbers := randomSample(len(mailList), 10)
		log.Println(numbers)

		var totalSkip int
		for k, v := range numbers {
			// If sender, receiver same address will be skip
			if mailList[v] == fromAddress {
				totalSkip++
				continue
			}
			token := tokenMap[mailList[v]]
			toAddress := mailList[v]
			result := RequestSend(fromAddress, token, toAddress)

			empData = append(empData, []string{strconv.FormatInt(int64(k+1), 10), fromAddress, toAddress, result})
		}

		// retry
		if totalSkip > 0 {
			// Random mails for send
			rand.NewSource(time.Now().UnixNano())
			numbers := randomSample(len(mailList), totalSkip)
			for _, v := range numbers {
				// If sender, receiver same address will be skip
				if mailList[v] == fromAddress {
					continue
				}
			}
		}
	}

	now := time.Now().Format("02-01-2006 15:04:05")
	fileName := "SendMail_result_" + now + ".csv"
	csvFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	for _, empRow := range empData {
		_ = csvwriter.Write(empRow)
	}
	csvwriter.Flush()
	csvFile.Close()
}

func RequestSend(fromAddress, fromToken, toAddress string) string {
	fmt.Printf("From: %s, To: %s\n", fromAddress, toAddress)
	if toAddress == fromAddress {
		return ""
	}

	if len(fromToken) == 0 {
		log.Println("token is empty")
		return ""
	}

	data := Payload{
		UploadOnly: false,
		IsDraft:    false,
		From: From{
			Address: fromAddress,
			Name:    "",
		},
		To: []To{
			{
				Name: "", Address: toAddress,
			},
		},
		Cc:          []Cc{},
		Bcc:         []Bcc{},
		Attachments: []Attachments{},
		Subject:     "Golang HHHHH",
		Date:        "",
		HTML:        "<p>Test semd</p>",
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Println("#1: ", err)
		// handle err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://ethermail.io/api/users/submit", body)
	if err != nil {
		// handle err
		log.Println("#2: ", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "i18n_redirected=en; webmail="+fromToken+"; afid=6357970844d160db88e635e8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
		log.Println("#3: ", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Output: ", string(b))
	return string(b)
}

func randomSample(n, k int) []int {
	if k > n {
		return nil
	}

	nums := make([]int, n)
	for i := range nums {
		nums[i] = i
	}

	rand.Shuffle(len(nums), func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})

	return nums[:k]
}

func MakeReadAll(mailbox, message, token string) *ReponseMakeReadAll {
	type Payload struct {
		Message string `json:"message"`
		Seen    bool   `json:"seen"`
	}

	data := Payload{
		Message: message,
		Seen:    true,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("PUT", "https://ethermail.io/api/mailboxes/"+mailbox+"/messages", body)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Authority", "ethermail.io")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("webmail=%s;", token))
	req.Header.Set("Origin", "https://ethermail.io")
	req.Header.Set("Referer", "https://ethermail.io/webmail/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	xResp := new(ReponseMakeReadAll)
	err = json.Unmarshal(b, &xResp)
	if err != nil {
		log.Println(err)
	}

	// log.Printf("MakeReadAll Success: %t, Updated: %d", xResp.Success, xResp.Updated)
	return xResp

}

// 639552040af700b73d6898a1
func GetAllMailBox(mailbox, token string) string {
	type Payload struct {
		Next     string `json:"next"`
		Previous string `json:"previous"`
		Mailbox  string `json:"mailbox"`
	}

	data := Payload{
		Mailbox: mailbox,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://ethermail.io/api/messages/search", body)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Authority", "ethermail.io")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("webmail=%s;", token))
	req.Header.Set("Origin", "https://ethermail.io")
	req.Header.Set("Referer", "https://ethermail.io/webmail/"+mailbox)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	xResp := new(GetAllMailBoxResponse)
	err = json.Unmarshal(b, &xResp)
	if err != nil {
		log.Println(err)
	}

	var result []string
	for _, v := range xResp.Results {
		result = append(result, strconv.Itoa(v.ID))
	}

	return strings.Join(result, ",")
}

type GetAllMailBoxResponse struct {
	Success bool `json:"success"`
	Total   int  `json:"total"`
	Page    int  `json:"page"`
	// PreviousCursor bool `json:"previousCursor"`
	// NextCursor     bool `json:"nextCursor"`
	Results []struct {
		ID      int    `json:"id"`
		Mailbox string `json:"mailbox"`
		Thread  string `json:"thread"`
		From    struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"from"`
		To []struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"to"`
		Cc          []interface{} `json:"cc"`
		Bcc         []interface{} `json:"bcc"`
		MessageID   string        `json:"messageId"`
		Subject     string        `json:"subject"`
		Date        time.Time     `json:"date"`
		Idate       time.Time     `json:"idate"`
		Attachments bool          `json:"attachments"`
		Size        int           `json:"size"`
		Seen        bool          `json:"seen"`
		Deleted     bool          `json:"deleted"`
		Flagged     bool          `json:"flagged"`
		Draft       bool          `json:"draft"`
		Answered    bool          `json:"answered"`
		Forwarded   bool          `json:"forwarded"`
		References  []interface{} `json:"references"`
		ContentType struct {
			Value  string `json:"value"`
			Params struct {
				Protocol string `json:"protocol"`
				Boundary string `json:"boundary"`
			} `json:"params"`
		} `json:"contentType"`
		Encrypted bool `json:"encrypted"`
	} `json:"results"`
}

// s%3ABb7Px1MUwHoLdW7K1o6zVOjVEWJEioI-.yUrR73vhFrkAfTFfxSXRWdjWDZy%2BvoPYAI%2BwpknjVlk
func GetMailBoxes(token string) *GetMailBoxesResponse {
	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

	// curl 'https://ethermail.io/api/mailboxes' \
	//   -H 'authority: ethermail.io' \
	//   -H 'accept: */*' \
	//   -H 'cookie: i18n_redirected=en; _gcl_au=1.1.1811535258.1680069131; _gaexp=GAX1.2.62TFBaiVTeaqMR829T5ejg.19475.0; _gid=GA1.2.836493515.1680069131; _fbp=fb.1.1680069130817.1899186682; afid=6357970844d160db88e635e8; _rdt_uuid=1680069145672.779cf56a-3125-4803-9b69-8717047cd861; webmail=s%3ABb7Px1MUwHoLdW7K1o6zVOjVEWJEioI-.yUrR73vhFrkAfTFfxSXRWdjWDZy%2BvoPYAI%2BwpknjVlk; _hjSessionUser_2797136=eyJpZCI6Ijg5ZTQ2NDFhLTBjYzItNTcwMi05MWRlLWI4MTA1MjgzMjM5OCIsImNyZWF0ZWQiOjE2ODAwNjkxMzA3NTAsImV4aXN0aW5nIjp0cnVlfQ==; _gat_gtag_UA_217939158_1=1; _ga=GA1.1.293739611.1680069131; _ga_QP9D69XHDM=GS1.1.1680219875.6.1.1680219994.10.0.0' \
	//   -H 'referer: https://ethermail.io/webmail/' \
	//   --compressed

	req, err := http.NewRequest("GET", "https://ethermail.io/api/mailboxes", nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authority", "ethermail.io")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cookie", fmt.Sprintf("webmail=%s;", token))
	req.Header.Set("Referer", "https://ethermail.io/webmail/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	xResp := new(GetMailBoxesResponse)
	err = json.Unmarshal(b, &xResp)
	if err != nil {
		log.Println(err)
	}

	return xResp
}
