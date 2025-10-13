package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/russross/blackfriday/v2"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Attachment struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Content  string `json:"content"`
}

func parseMarkdown(s string) string {
	buf := bytes.NewBufferString(s)
	output := blackfriday.Run(buf.Bytes(), blackfriday.WithExtensions(blackfriday.HardLineBreak))
	return string(output)
}

func sendDebugEmail(subject string, body string) error {
	var err error

	// Init
	fromName := "VM Notifications"
	if "production" != os.Getenv("VM_ENVIRONMENT") {
		fromName += " (" + strings.ToUpper(os.Getenv("VM_ENVIRONMENT")) + ")"
	}

	from := mail.NewEmail(fromName, "noreply@vm.com")
	to := mail.NewEmail("Haresh Suralkar", "suralkar.haresh@gmail.com")
	message := mail.NewSingleEmail(from, subject, to, body, body)

	// Send
	client := sendgrid.NewSendClient(os.Getenv("VM_SENDGRID_API_KEY"))
	_, err = client.Send(message)

	return err
}

func (e *EmailTemplate) buildEmailTemplateBody() string {
	body := parseMarkdown(e.Body)

	b, err := ioutil.ReadFile("data/email_template.html")
	if err == nil {
		css, err := ioutil.ReadFile("data/emails.css")
		if err == nil {
			urlBase := os.Getenv("VM_EMAIL_ASSETS_URL")

			file := string(b)
			file = strings.Replace(file, ":css", string(css), 1)

			file = strings.Replace(file, ":header_src", "header-v2.jpg", 1)

			file = strings.Replace(file, ":preheader", e.Preheader, 1)
			file = strings.Replace(file, ":subject", e.Subject, 1)
			file = strings.Replace(file, ":body", body, 1)

			file = strings.Replace(file, "src=\"", "src=\""+urlBase, -1) // 'src="' . get_stylesheet_directory_uri() . '/emails/',

			body = file

			if e.Preheader == "" {
				re := regexp.MustCompile("(?s)<!-- preheader -->(.+?)<!-- /preheader -->")
				body = re.ReplaceAllString(body, "")
			}
		}
	}

	return body
}

func sendEmail(language string, slug string, mvs map[string]string, attachments []Attachment) error {
	var t EmailTemplate

	email, ok := mvs["to_email"]
	if !ok {
		return errors.New("Email address not set")
	}

	name, ok := mvs["full_name"]
	if !ok {
		name = email
	}

	// Check email address
	// if !allowSendingtoEmail(email) {
	// 	return errors.New("Email domain is not whitelisted")
	// }

	if language == "" {
		language = "en"
	}

	if db.Where("language = ?", language).Where("slug = ?", slug).First(&t).RecordNotFound() {
		return errors.New("Template not found")
	}

	// Check valid template
	if t.Subject == "" || t.Body == "" || t.FromName == "" || t.FromEmail == "" {
		return errors.New("Template is invalid")
	}

	all, err := json.Marshal(mvs)
	if err == nil {
		mvs["all"] = string(all)
	}

	mvs["profile_url"] = os.Getenv("VM_PROFILE_URL")

	// Apply merge mvs
	for find, replace := range mvs {
		t.FromName = strings.Replace(t.FromName, ":"+find, replace, -1)
		t.Subject = strings.Replace(t.Subject, ":"+find, replace, -1)
		t.Preheader = strings.Replace(t.Preheader, ":"+find, replace, -1)
		t.Body = strings.Replace(t.Body, ":"+find, replace, -1)
	}

	// Init
	fromName := t.FromName
	if "production" != os.Getenv("VM_ENVIRONMENT") {
		fromName += " (" + strings.ToUpper(os.Getenv("VM_ENVIRONMENT")) + ")"
	}

	from := mail.NewEmail(fromName, t.FromEmail)
	to := mail.NewEmail(name, email)
	message := mail.NewSingleEmail(from, t.Subject, to, t.Body, t.buildEmailTemplateBody())

	// Attachments
	if len(attachments) > 0 {
		for _, a := range attachments {
			att := mail.NewAttachment()
			att.SetContent(a.Content)
			att.SetType(a.Type)
			att.SetFilename(a.Filename)
			message.AddAttachment(att)
		}
	}

	// Send
	client := sendgrid.NewSendClient(os.Getenv("VM_SENDGRID_API_KEY"))
	_, err = client.Send(message)

	return err
}

func (u *User) notify(slug string, email string, mvs map[string]string, attachments []Attachment) error {
	// Set merge mvs
	if mvs == nil {
		mvs = make(map[string]string)
	}

	mvs["to_email"] = email
	mvs["full_name"] = u.getName()
	mvs["first_name"] = u.FirstName
	mvs["last_name"] = u.LastName

	return sendEmail(u.Language, slug, mvs, attachments)
}
