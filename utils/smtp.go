package utils

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net/smtp"
	"strconv"

	"chinese-chess-backend/config"
)

func GetRandomCode() string {
	const min = 100000
	const max = 999999

	// 使用 crypto/rand 生成4个随机字节
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		log.Fatal("生成验证码失败:", err)
	}
	// 将随机结果转换为正整数，并计算出对应范围内的数字
	n := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if n < 0 {
		n = -n
	}
	code := n%(max-min+1) + min
	return strconv.Itoa(code)
}

func SendMail(to, subject, body string) error {
	smtpConfig := config.GetSMTPConfig()
	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)
	msg := "From: " + smtpConfig.Username + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body
	addr := smtpConfig.Host + ":" + smtpConfig.Port

	// Establishing TLS connection
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpConfig.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		log.Printf("Error establishing TLS connection: %v", err)
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpConfig.Host)
	if err != nil {
		log.Printf("Error creating SMTP client: %v", err)
		return err
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		log.Printf("Error during SMTP authentication: %v", err)
		return err
	}

	if err = client.Mail(smtpConfig.Username); err != nil {
		log.Printf("Error setting sender: %v", err)
		return err
	}

	if err = client.Rcpt(to); err != nil {
		log.Printf("Error setting recipient: %v", err)
		return err
	}

	w, err := client.Data()
	if err != nil {
		log.Printf("Error getting data writer: %v", err)
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		log.Printf("Error writing message: %v", err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Printf("Error closing writer: %v", err)
		return err
	}

	log.Println("Email sent successfully")
	return nil
}