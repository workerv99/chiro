package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
)

var (
	apiKey    string
	fromEmail string
)

func Init() {
	apiKey = os.Getenv("RESEND_API_KEY")
	fromEmail = os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "noreply@chiro.app"
	}
}

func IsConfigured() bool {
	return apiKey != ""
}

type EmailData struct {
	Name    string
	Action  string
	URL     string
	Code    string
	Year    int
}

func SendWelcome(to, name string) error {
	return send(to, "Bienvenido a Chiro", "welcome", EmailData{Name: name, Action: "welcome", Year: 2026})
}

func SendPasswordReset(to, name, resetURL string) error {
	return send(to, "Restablece tu contraseña", "reset", EmailData{Name: name, Action: "reset", URL: resetURL, Year: 2026})
}

func SendVerifyEmail(to, name, verifyURL string) error {
	return send(to, "Verifica tu email", "verify", EmailData{Name: name, Action: "verify", URL: verifyURL, Year: 2026})
}

func send(to, subject, templateName string, data EmailData) error {
	if !IsConfigured() {
		return fmt.Errorf("email no configurado")
	}

	htmlBody, err := renderTemplate(templateName, data)
	if err != nil {
		return err
	}

	payload := fmt.Sprintf(`{
		"from": "%s",
		"to": ["%s"],
		"subject": "%s",
		"html": %q
	}`, fromEmail, to, subject, htmlBody)

	req, _ := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBufferString(payload))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("error enviando email: status %d", resp.StatusCode)
	}
	return nil
}

func renderTemplate(name string, data EmailData) (string, error) {
	tmpl := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:480px;margin:0 auto;padding:32px">
<div style="text-align:center;margin-bottom:24px">
<div style="width:48px;height:48px;border-radius:12px;background:#f1f5f9;display:inline-flex;align-items:center;justify-content:center;font-weight:800;color:#5b7cf6;font-size:20px">C</div>
</div>
<h1 style="font-size:20px;font-weight:700;text-align:center;margin-bottom:16px">{{.Action}}</h1>
<p style="color:#64748b;text-align:center;margin-bottom:24px">Hola {{.Name}},</p>
{{if .URL}}
<div style="text-align:center;margin:24px 0">
<a href="{{.URL}}" style="background:#5b7cf6;color:#fff;padding:12px 24px;border-radius:8px;text-decoration:none;font-weight:600;display:inline-block">Continuar</a>
</div>
{{end}}
<p style="color:#94a3b8;font-size:12px;text-align:center;margin-top:32px">© {{.Year}} Chiro. Todos los derechos reservados.</p>
</body>
</html>`

	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	return buf.String(), err
}
