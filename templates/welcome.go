package templates

import (
	"fmt"
	"shared/constants"
	"time"

	"github.com/matcornic/hermes/v2"
)

func GenerateWelcome(fullName string, loginURL string, username string, tempPassword string, companyName string, licenseKey string) (string, error) {
	copyrigt := fmt.Sprintf("© %s Meizo Infotech Private Limited. All rights reserved.", time.Now().Format("2006"))

	h := hermes.Hermes{
		Product: hermes.Product{
			Name:      "Meizo HRMS",
			Link:      "https://meizoerp.com",
			Logo:      "", // optional: set logo URL if you have one
			Copyright: copyrigt,
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: fullName,
			Intros: []string{
				"Welcome to Meizo HRMS! 🎉",
				"We're thrilled to have you onboard. Below are your login details to access the HRMS platform.",
			},
			Dictionary: []hermes.Entry{
				{Key: "Login URL", Value: loginURL},
				{Key: "Username", Value: username},
				{Key: "Temporary Password", Value: tempPassword},
				{Key: "License key", Value: licenseKey},
			},
			Outros: []string{
				"Please log in using these credentials and make sure to reset your password after your first login.",
				"If you need help or have any questions, feel free to reach out to the support team.",
				"at " + constants.FromEmail,
				"Best regards,",
			},
			Signature: `Meizo Infotech Private Limited`,
		},
	}

	return h.GenerateHTML(email)
}

func GenerateRevokeEmail(email string, name string) (string, error) {
	copyrigt := fmt.Sprintf("© %s Meizo Infotech Private Limited. All rights reserved.", time.Now().Format("2006"))
	h := hermes.Hermes{
		Product: hermes.Product{
			Name:      "Meizo HRMS",
			Link:      "https://meizoerp.com",
			Logo:      "", // optional: set logo URL if you have one
			Copyright: copyrigt,
		},
	}

	emailContent := hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				"Access to Meizo HRMS has been revoked",
			},
			Outros: []string{
				"Your access to Meizo HRMS has been revoked and you cannot access the any of the services of the hrms.",
				"If you need help or have any questions, feel free to reach out to the support team.",
				"at " + constants.FromEmail,
				"Best regards,",
			},
			Signature: `Meizo Infotech Private Limited`,
		},
	}

	return h.GenerateHTML(emailContent)
}

// GenerateEmployeeWelcome generates an employee-specific welcome email template
func GenerateEmployeeWelcome(fullName string, firstName string, lastName string, employeeCode string, tempPassword string, companyCode string, designation string, department string) (string, error) {
	copyrigt := fmt.Sprintf("© %s Meizo Infotech Private Limited. All rights reserved.", time.Now().Format("2006"))

	h := hermes.Hermes{
		Product: hermes.Product{
			Name:      "Meizo HRMS",
			Link:      "https://meizoerp.com",
			Logo:      "", // optional: set logo URL if you have one
			Copyright: copyrigt,
		},
	}

	email := hermes.Email{
		Body: hermes.Body{
			Name: fullName,
			Intros: []string{
				"Welcome to Meizo HRMS Employee Self-Service Portal! 🎉",
				"We're excited to have you as part of our team. Your employee portal access has been created.",
			},
			Dictionary: []hermes.Entry{
				{Key: "Employee Code", Value: employeeCode},
				{Key: "Name", Value: fullName},
				{Key: "Designation", Value: designation},
				{Key: "Department", Value: department},
				{Key: "Company Code", Value: companyCode},
				{Key: "Username", Value: employeeCode},
				{Key: "Temporary Password", Value: tempPassword},
			},
			Outros: []string{
				"Please log in using your Employee Code as username and the temporary password provided above.",
				"After your first login, you will be prompted to change your password for security purposes.",
				"You will have access to the Employee Self-Service (ESS) module where you can:",
				"• View and update your personal information",
				"• Access your payslips and tax documents",
				"• Apply for leave and view attendance",
				"• Update your profile and more",
				"If you need help or have any questions, please contact your HR department or reach out to support at " + constants.FromEmail,
				"Best regards,",
			},
			Signature: `Meizo Infotech Private Limited`,
		},
	}

	return h.GenerateHTML(email)
}
