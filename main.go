package main

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"text/template"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

// Error handling for GetProjectRoot
func GetProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current working directory:", err)
	}
	return cwd
}

// Refactored template rendering function
func renderTemplate(w http.ResponseWriter, tmplName string) {
	templatePath := filepath.Join(GetProjectRoot(), tmplName)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	// No need for RouteChecker middleware
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux, // Directly using the mux without middleware
	}

	fmt.Println("server running @http://localhost:8080\n=====================================")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error running server: ", err)
	}
}

func RegisterRoutes(mux *http.ServeMux) {
	staticDir := GetProjectRoot()

	// Serve static assets (css, js, img, lib) from the correct folder
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(filepath.Join(staticDir, "css")))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(filepath.Join(staticDir, "js")))))
	mux.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(filepath.Join(staticDir, "img")))))
	mux.Handle("/favicon/", http.StripPrefix("/favicon/", http.FileServer(http.Dir(filepath.Join(staticDir, "favicon")))))
	// mux.Handle("/scss/", http.StripPrefix("/scss/", http.FileServer(http.Dir(filepath.Join(staticDir, "scss")))))

	// Register the home page handler and email handler
	mux.HandleFunc("/", HomeHandler)
	mux.HandleFunc("/resume", handleresume)
	mux.HandleFunc("/send-email", handleEmailSend)
}

// HomeHandler handles the home page request
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "index.html")
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Contact page handler
func handleresume(w http.ResponseWriter, r *http.Request) {
	// Path to the PDF file
	filePath := "resume/SAMUELOKOTHOMULORESUME.pdf"

	// Set the appropriate headers for PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\"SAMUELOKOTHOMULORESUME.pdf\"")

	// Serve the PDF file
	http.ServeFile(w, r, filePath)
}

// // Contact page handler
// func handleContact(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "contact.html")
// }

// // Contact page handler
// func handleaward(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "award.html")
// }

// // About page handler
// func handleAbout(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "about.html")
// }

// // Blogs page handler
// func handleBlogs(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "blogs.html")
// }

// // Experience page handler
// func handleExperience(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "experience.html")
// }

// // My Work page handler
// func handleMyWork(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "mywork.html")
// }

// // Skills page handler
// func handleSkills(w http.ResponseWriter, r *http.Request) {
// 	renderTemplate(w, "skills.html")
// }

// // Testimonial page handler
//
//	func handleTestimonial(w http.ResponseWriter, r *http.Request) {
//		renderTemplate(w, "testimonial.html")
//	}
func handleEmailSend(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Initialize the bluemonday policy (sanitize HTML content)
	p := bluemonday.UGCPolicy()

	// Sanitize form fields
	name := p.Sanitize(r.FormValue("name"))
	email := p.Sanitize(r.FormValue("email"))
	subject := p.Sanitize(r.FormValue("subject"))
	message := p.Sanitize(r.FormValue("message"))

	// Validate Name (letters, spaces, hyphens, apostrophes, accents)
	nameRegex := `^[a-zA-Z\sÀ-ÿ'-]+$`
	if !regexp.MustCompile(nameRegex).MatchString(name) {
		http.Error(w, "Invalid name format", http.StatusBadRequest)
		return
	}

	// Validate Email
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if !regexp.MustCompile(emailRegex).MatchString(email) {
		http.Error(w, "Invalid email address format", http.StatusBadRequest)
		return
	}

	// Validate Subject (Allow more characters but with reasonable length)
	if len(subject) < 2 || len(subject) > 100 {
		http.Error(w, "Subject must be between 2-100 characters", http.StatusBadRequest)
		return
	}

	// Validate Message (Allow more types of characters for better communication)
	if len(message) < 10 || len(message) > 2000 {
		http.Error(w, "Message must be between 10-2000 characters", http.StatusBadRequest)
		return
	}

	// Check if all fields are provided
	if name == "" || email == "" || subject == "" || message == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Configure email settings
	from := "mcomulosammy37@gmail.com"
	password := "eptw taih ilqn srls" // Consider using environment variables for this
	to := "mcomulosammy37@gmail.com"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Record the timestamp for consistent use
	currentTime := time.Now()
	formattedDate := currentTime.Format("January 2, 2006")

	// Compose the admin notification email
	adminEmailSubject := "New Portfolio Inquiry: " + subject
	adminEmailBody := fmt.Sprintf(`
<html>
<head>
<title>New Portfolio Inquiry</title>
<style>
body {
  font-family: Arial, sans-serif;
  color: rgb(15, 15, 15);
  margin: 0;
  padding: 0;
}
.container {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
  background: #f8f8f8;
  border-radius: 8px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}
.header {
  background-color: #2196f3;
  color: white;
  padding: 15px 20px;
  text-align: center;
  border-radius: 8px 8px 0 0;
}
.footer {
  background-color: #f1f1f1;
  padding: 15px;
  text-align: center;
  margin-top: 20px;
  font-size: 12px;
  color: #777;
  border-radius: 0 0 8px 8px;
}
.content {
  padding: 25px;
  background-color: white;
  border-radius: 4px;
  margin: 15px 0;
}
.content p {
  font-size: 16px;
  line-height: 1.6;
  margin: 10px 0;
}
.metadata {
  background-color: #f9f9f9;
  padding: 10px 15px;
  border-radius: 4px;
  margin-top: 20px;
  font-size: 13px;
  color: #666;
}
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h2>New Portfolio Inquiry</h2>
  </div>
  <div class="content">
    <p><strong>Name:</strong> %s</p>
    <p><strong>Email:</strong> %s</p>
    <p><strong>Subject:</strong> %s</p>
    <p><strong>Message:</strong></p>
    <p style="background-color: #f9f9f9; padding: 15px; border-left: 4px solid #2196f3; margin: 15px 0;">%s</p>
  </div>
  <div class="metadata">
    <p>Received: %s</p>
    <p>IP Address: %s (for security monitoring)</p>
  </div>
  <div class="footer">
    <p>&copy; %d Samuel Omulo Portfolio. All rights reserved.</p>
  </div>
</div>
</body>
</html>
`, name, email, subject, message, formattedDate, r.RemoteAddr, currentTime.Year())

	// Compose the client confirmation email
	clientEmailSubject := "Thank you for contacting Samuel Omulo"
	clientEmailBody := fmt.Sprintf(`
<html>
<head>
<title>Thank You for Your Message</title>
<style>
body {
  font-family: Arial, sans-serif;
  color: #333;
  margin: 0;
  padding: 0;
}
.container {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
  background: #f8f8f8;
  border-radius: 8px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}
.header {
  background-color: #2196f3;
  color: white;
  padding: 15px 20px;
  text-align: center;
  border-radius: 8px 8px 0 0;
}
.logo {
  max-width: 100px;
  margin: 0 auto 15px;
  display: block;
}
.footer {
  background-color: #f1f1f1;
  padding: 15px;
  text-align: center;
  margin-top: 20px;
  font-size: 12px;
  color: #777;
  border-radius: 0 0 8px 8px;
}
.content {
  padding: 25px;
  background-color: white;
  border-radius: 4px;
  margin: 15px 0;
  line-height: 1.6;
}
.message-summary {
  background-color: #f9f9f9;
  padding: 15px;
  border-radius: 4px;
  margin: 20px 0;
}
.signature {
  margin-top: 30px;
  padding-top: 15px;
  border-top: 1px solid #eee;
}
.social-links {
  margin-top: 15px;
}
.social-links a {
  color: #2196f3;
  margin: 0 5px;
  text-decoration: none;
}
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h2>Thank You for Your Message</h2>
  </div>
  <div class="content">
    <p>Dear %s,</p>
    
    <p>Thank you for reaching out through my portfolio website. I have received your message regarding <strong>"%s"</strong> and appreciate your interest in my services.</p>
    
    <p>I will review your inquiry and respond as soon as possible, typically within 24-48 hours during business days.</p>
    
    <div class="message-summary">
      <p><strong>Your message details:</strong></p>
      <p><strong>Date:</strong> %s</p>
      <p><strong>Subject:</strong> %s</p>
      <p><strong>Message:</strong> %s</p>
    </div>
    
    <p>If you have any urgent matters or additional information to share, please feel free to reply to this email.</p>
    
    <div class="signature">
      <p>Best regards,</p>
      <p><strong>Samuel Omulo</strong></p>
      <p>Professional Software Developer</p>
      <div class="social-links">
        <a href="https://linkedin.com/in/yourprofile">LinkedIn</a> | 
        <a href="https://github.com/somulo1">GitHub</a>
      </div>
    </div>
  </div>
  <div class="footer">
    <p>&copy; %d Samuel Omulo. All rights reserved.</p>
  </div>
</div>
</body>
</html>
`, name, subject, formattedDate, subject, message, currentTime.Year())

	// Set up authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send email to admin (yourself)
	adminMsg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		to, adminEmailSubject, adminEmailBody))

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, adminMsg)
	if err != nil {
		log.Printf("Error sending admin notification email: %v", err)
		http.Error(w, "Failed to process your request", http.StatusInternalServerError)
		return
	}

	// Send confirmation email to the client
	clientMsg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		email, clientEmailSubject, clientEmailBody))

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{email}, clientMsg)
	if err != nil {
		log.Printf("Error sending client confirmation email: %v", err)
		// Continue even if client email fails, as we already received the message
	}

	// Return a professional confirmation page to the user
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
	<title>Message Submitted Successfully</title>
	<link rel="stylesheet" href="css/popup_submeet.css">
	<style>
		.confirmation {
			max-width: 600px;
			margin: 50px auto;
			padding: 30px;
			background: #fff;
			border-radius: 8px;
			box-shadow: 0 5px 15px rgba(255, 255, 255, 0.1);
			text-align: center;
		}
		p {
			font-size: 18px;
			color: rgb(0, 0, 0);
			}
		p1 {
			font-size: 18px;
			color: rgb(255, 255, 255);
			}	
		h2 {
			color: #2196f3;
			margin-bottom: 20px;
		}
		.message-details {
			background:rgb(0, 0, 0);
			padding: 20px;
			border-radius: 6px;
			margin: 20px 0;
			text-align: left;
		}
		.message-details p {
			margin: 10px 0;
		}
		.success-icon {
			font-size: 60px;
			color: #4CAF50;
			margin-bottom: 20px;
		}
		.back-button {
			background: #2196f3;
			color: white;
			border: none;
			padding: 12px 24px;
			border-radius: 4px;
			cursor: pointer;
			font-size: 16px;
			margin-top: 20px;
			transition: background 0.3s;
		}
		.back-button:hover {
			background: #0b7dda;
		}
		.footer {
			margin-top: 30px;
			font-size: 14px;
			color: #777;
		}
	</style>
	</head>
	<body>
	<div class="confirmation">
		<div class="success-icon">✓</div>
		<h2>Message Sent Successfully!</h2>
		<p>Thank you for reaching out, %s. I've received your message and will respond shortly.</p>
		<p>A confirmation email has been sent to <strong>%s</strong>.</p>
		
		<div class="message-details">
			<p1><strong>Subject:</strong> %s</p1>
			<p1><strong>Submitted:</strong> %s</p1>
		</div>
		
		<button class="back-button" onclick="window.location.href='/'">Return to Portfolio</button>
		
		<div class="footer">
			<p>&copy; %d Samuel Omulo. All rights reserved.</p>
		</div>
	</div>
	</body>
	</html>
	`, name, email, subject, formattedDate, currentTime.Year())
}
