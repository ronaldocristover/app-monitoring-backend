package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ronaldocristover/app-monitoring/internal/model"
)

func main() {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5433"
	}
	if user == "" {
		user = "app"
	}
	if pass == "" {
		pass = "secret"
	}
	if dbname == "" {
		dbname = "app_monitoring"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	users := []model.User{
		{Name: "Ronaldo Cristover", Email: "ronaldo@example.com", PasswordHash: string(passwordHash)},
		{Name: "Andi Pratama", Email: "andi@example.com", PasswordHash: string(passwordHash)},
		{Name: "Siti Nurhaliza", Email: "siti@example.com", PasswordHash: string(passwordHash)},
		{Name: "Budi Santoso", Email: "budi@example.com", PasswordHash: string(passwordHash)},
		{Name: "Dewi Lestari", Email: "dewi@example.com", PasswordHash: string(passwordHash)},
	}
	for i := range users {
		db.Create(&users[i])
	}
	fmt.Printf("Users: %d\n", len(users))

	apps := []model.App{
		{AppName: "LMS Backend", Description: "Learning Management System", Tags: "education,internal"},
		{AppName: "GBA Youth Portal", Description: "Youth empowerment platform", Tags: "youth,community"},
		{AppName: "HKCCC Website", Description: "Hong Kong CCC main site", Tags: "corporate,public"},
		{AppName: "Stage Edu Net", Description: "Education network platform", Tags: "education,network"},
		{AppName: "Upower App", Description: "Power management dashboard", Tags: "energy,internal"},
		{AppName: "InnerLab CMS", Description: "Content management system", Tags: "cms,internal"},
		{AppName: "E-Commerce API", Description: "Online shop backend service", Tags: "commerce,public"},
	}
	for i := range apps {
		db.Create(&apps[i])
	}
	fmt.Printf("Apps: %d\n", len(apps))

	servers := []model.Server{
		{Name: "Geboy Dev", IP: "159.65.11.160", Provider: "DigitalOcean"},
		{Name: "Geboy Prod", IP: "188.166.182.236", Provider: "DigitalOcean"},
		{Name: "HK Primary", IP: "103.75.201.50", Provider: "AWS"},
		{Name: "SG Staging", IP: "128.199.95.100", Provider: "DigitalOcean"},
		{Name: "ID On-Premise", IP: "192.168.1.100", Provider: "On-Premise"},
	}
	for i := range servers {
		db.Create(&servers[i])
	}
	fmt.Printf("Servers: %d\n", len(servers))

	envNames := []string{"development", "staging", "production"}
	var environments []model.Environment
	for _, app := range apps {
		numEnvs := 1 + rand.Intn(3)
		for j := 0; j < numEnvs; j++ {
			env := model.Environment{AppID: app.ID, Name: envNames[j%3]}
			db.Create(&env)
			environments = append(environments, env)
		}
	}
	fmt.Printf("Environments: %d\n", len(environments))

	serviceTypes := []string{"backend", "frontend", "worker", "cron", "api-gateway"}
	frameworks := []string{"NestJS", "Go/Gin", "Next.js", "Express", "FastAPI"}
	dbTypes := []string{"PostgreSQL", "MySQL", "MongoDB", "Redis", "SQLite"}
	var services []model.Service
	for i, env := range environments {
		if i >= 21 {
			break
		}
		svc := model.Service{
			EnvironmentID:  env.ID,
			ServerID:       servers[rand.Intn(len(servers))].ID,
			Name:           fmt.Sprintf("%s-%s", apps[i%7].AppName, serviceTypes[i%5]),
			Type:           serviceTypes[i%5],
			URL:            fmt.Sprintf("https://%s.example.com", serviceTypes[i%5]),
			Repository:     fmt.Sprintf("https://github.com/ronaldocristover/%s", serviceTypes[i%5]),
			StackLanguage:  "Go",
			StackFramework: frameworks[i%5],
			DBType:         dbTypes[i%5],
			DBHost:         servers[rand.Intn(len(servers))].IP,
		}
		db.Create(&svc)
		services = append(services, svc)
	}
	fmt.Printf("Services: %d\n", len(services))

	for _, svc := range services {
		enabled := rand.Intn(10) > 1
		cfg := model.MonitoringConfig{
			ServiceID:           svc.ID,
			Enabled:             enabled,
			PingIntervalSeconds: 30 + rand.Intn(5)*30,
			TimeoutSeconds:      10,
			Retries:             3,
		}
		db.Create(&cfg)
	}
	fmt.Printf("MonitoringConfigs: %d\n", len(services))

	now := time.Now()
	for i := 0; i < 51; i++ {
		svc := services[rand.Intn(len(services))]
		status := "up"
		statusCode := 200
		errMsg := ""
		if rand.Intn(10) == 0 {
			status = "down"
			statusCode = 503
			errMsg = "connection timeout"
		}
		ml := model.MonitoringLog{
			ServiceID:      svc.ID,
			Status:         status,
			ResponseTimeMs: 20 + rand.Intn(480),
			StatusCode:     statusCode,
			ErrorMessage:   errMsg,
			CheckedAt:      now.Add(-time.Duration(i*5) * time.Minute),
		}
		db.Create(&ml)
	}
	fmt.Printf("MonitoringLogs: 51\n")

	methods := []string{"docker", "git-pull", "helm", "manual"}
	for i := 0; i < 13; i++ {
		svc := services[rand.Intn(len(services))]
		dep := model.Deployment{
			ServiceID:     svc.ID,
			Method:        methods[i%4],
			ContainerName: fmt.Sprintf("%s-container", svc.Name),
			Port:          3000 + i,
			Config:        fmt.Sprintf(`{"replicas": %d, "auto_scale": true}`, 1+i%3),
		}
		db.Create(&dep)
	}
	fmt.Printf("Deployments: 13\n")

	schedules := []string{"0 2 * * *", "0 3 * * 0", "0 0 1 * *"}
	backupStatuses := []string{"success", "pending", "failed"}
	for i := 0; i < 7; i++ {
		svc := services[rand.Intn(len(services))]
		t := now.Add(-time.Duration(i*12) * time.Hour)
		bk := model.Backup{
			ServiceID:      svc.ID,
			Enabled:        true,
			Path:           fmt.Sprintf("/backups/%s/", svc.Name),
			Schedule:       schedules[i%3],
			LastBackupTime: &t,
			Status:         backupStatuses[i%3],
		}
		db.Create(&bk)
	}
	fmt.Printf("Backups: 7\n")
	fmt.Println("Seed complete!")
}
