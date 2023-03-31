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
	// open file
	f, err := os.Open("Ethermail_List.csv") // headers No.,	Ethermail Address,	Token at row 1
	if err != nil {
		log.Fatal(err)
	}

	// close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// convert csv data to map
	tokenMap := make(map[string]string)
	for _, line := range csvData {
		mail := line[1]
		token := line[2]
		tokenMap[mail] = token
	}

	// Getting a slice
	mailList := []string{}
	for k := range tokenMap {
		mailList = append(mailList, k)
	}

	fmt.Println("Service starting...")

	// Set timezone
	location, _ := time.LoadLocation("Asia/Bangkok")

	// New scheduler with timezone
	scheduler := gocron.NewScheduler(location)
	scheduler.Every(1).Day().At("11:59").Do(func() {

		fmt.Println("Starting Send mails")
		SendMail(mailList, tokenMap)
		fmt.Println("End Send mails")

		fmt.Println("Starting Read messages")
		ReadAllMessage(mailList, tokenMap)
		fmt.Println("End Read messages")
	})

	// start scheduler
	scheduler.StartAsync()
	scheduler.StartBlocking()
}

func ReadAllMessage(mailList []string, tokenMap map[string]string) {
	// Csv Data
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

		// Added csv data
		empData = append(empData, []string{strconv.FormatInt(int64(k+1), 10), mail, mailIDs, strconv.FormatBool(resulrMakeRead.Success), strconv.Itoa(resulrMakeRead.Updated), remark})
	}

	// filename
	now := time.Now().Format("02-01-2006 15:04:05")
	fileName := "ReadAllMessage_result_" + now + ".csv"

	// Create csv file
	csvFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	// Write csv file
	csvwriter := csv.NewWriter(csvFile)
	for _, empRow := range empData {
		_ = csvwriter.Write(empRow)
	}
	csvwriter.Flush()
	csvFile.Close()
}

func SendMail(mailList []string, tokenMap map[string]string) {
	// Csv Data
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

			// Added csv data
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

	// filename
	now := time.Now().Format("02-01-2006 15:04:05")
	fileName := "SendMail_result_" + now + ".csv"

	csvFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	// Create csv file
	csvwriter := csv.NewWriter(csvFile)
	for _, empRow := range empData {
		_ = csvwriter.Write(empRow)
	}

	// Write csv file
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
		Subject:     "Hi Bigkz,",
		Date:        "",
		HTML: `
		<p>How's it going?</p>
		<p>How's it going?</p>
		<p>Sorry I haven't been in touch for such a long time but I've had exams so I've been studying every free minute. Anyway, I'd love to hear all your news and I'm hoping we can get together soon to catch up. We just moved to a bigger flat so maybe you can come and visit one weekend?</p>
		<p>How's the new job? &nbsp;</p>
		<p>Looking forward to hearing from you!</p>
		<p>Helga</p>
		<hr><p>Hi Helga,</p>
		<p>I've been meaning to write to you for ages now so don't worry! How did your exams go? When will you know your results? I'm sure you did brilliantly as always!</p>
		<p>As for me, I'll have been in the new job three months by the end of next week so I'm feeling more settled in. At first I felt like I had no idea what I was doing but now I realise it's normal to feel like that. There was a lot to learn – there still is actually – and I soon had to get used to the idea that I can't know everything. I used to work late a lot and at weekends but I'm slowly getting into a normal routine.</p>
		<p>Which means I'd love to come and visit! We really need a good catch up! I can't believe we haven't seen each other since Carl's wedding. How does next month sound?</p>
		<p>Anyway, I'd better get back to work.</p>
		<p>Congratulations on the new flat!&nbsp;Can't wait to see you!</p>
		<p>Love,</p>
		<p>Bigkz</p>
		`,
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
	req.Header.Set("Cookie", fmt.Sprintf("webmail=%s;", fromToken))

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

	return xResp
}

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

func GetMailBoxes(token string) *GetMailBoxesResponse {

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
