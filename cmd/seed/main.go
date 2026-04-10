package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ronaldocristover/app-monitoring/internal/config"
	"github.com/ronaldocristover/app-monitoring/internal/model"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.User, cfg.Database.Password, cfg.Database.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	// --- 1. Users (5) ---
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	users := []model.User{
		{ID: uuid.New(), Name: "Ronaldo Cristover", Email: "ronaldo@app-monitoring.com", PasswordHash: string(passwordHash)},
		{ID: uuid.New(), Name: "Andi Pratama", Email: "andi@app-monitoring.com", PasswordHash: string(passwordHash)},
		{ID: uuid.New(), Name: "Siti Nurhaliza", Email: "siti@app-monitoring.com", PasswordHash: string(passwordHash)},
		{ID: uuid.New(), Name: "Budi Santoso", Email: "budi@app-monitoring.com", PasswordHash: string(passwordHash)},
		{ID: uuid.New(), Name: "Dewi Lestari", Email: "dewi@app-monitoring.com", PasswordHash: string(passwordHash)},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			log.Printf("Skip user %s: %v", users[i].Email, err)
		}
	}
	fmt.Printf("✅ Seeded %d users\n", len(users))

	// --- 2. Apps (7) ---
	apps := []model.App{
		{ID: uuid.New(), AppName: "LMS Backend", Description: "Learning Management System", Tags: "education,internal"},
		{ID: uuid.New(), AppName: "GBA Youth Portal", Description: "Youth empowerment platform", Tags: "youth,community"},
		{ID: uuid.New(), AppName: "HKCCC Website", Description: "Hong Kong CCC main site", Tags: "corporate,public"},
		{ID: uuid.New(), AppName: "Stage Edu Net", Description: "Education network platform", Tags: "education,network"},
		{ID: uuid.New(), AppName: "Upower App", Description: "Power management dashboard", Tags: "energy,internal"},
		{ID: uuid.New(), AppName: "InnerLab CMS", Description: "Content management system", Tags: "cms,internal"},
		{ID: uuid.New(), AppName: "E-Commerce API", Description: "Online shop backend service", Tags: "commerce,public"},
	}
	for i := range apps {
		db.Create(&apps[i])
	}
	fmt.Printf("✅ Seeded %d apps\n", len(apps))

	// --- 3. Servers (5) ---
	servers := []model.Server{
		{ID: uuid.New(), Name: "Geboy Dev", IP: "159.65.11.160", Provider: "DigitalOcean"},
		{ID: uuid.New(), Name: "Geboy Prod", IP: "188.166.182.236", Provider: "DigitalOcean"},
		{ID: uuid.New(), Name: "HK Primary", IP: "103.75.201.50", Provider: "AWS"},
		{ID: uuid.New(), Name: "SG Staging", IP: "128.199.95.100", Provider: "DigitalOcean"},
		{ID: uuid.New(), Name: "ID On-Premise", IP: "192.168.1.100", Provider: "On-Premise"},
	}
	for i := range servers {
		db.Create(&servers[i])
	}
	fmt.Printf("✅ Seeded %d servers\n", len(servers))

	// --- 4. Environments (15) ---
	envNames := []string{"development", "staging", "production"}
	var environments []model.Environment
	for _, app := range apps {
		numEnvs := 1 + rand.Intn(3) // 1-3 envs per app
		for j := 0; j < numEnvs; j++ {
			env := model.Environment{
				ID:     uuid.New(),
				AppID:  app.ID,
				Name:   envNames[j%3],
			}
			db.Create(&env)
			environments = append(environments, env)
		}
	}
	fmt.Printf("✅ Seeded %d environments\n", len(environments))

	// --- 5. Services (21) ---
	serviceTypes := []string{"backend", "frontend", "worker", "cron", "api-gateway"}
	frameworks := []string{"NestJS", "Go/Gin", "Next.js", "Express", "FastAPI"}
	dbTypes := []string{"PostgreSQL", "MySQL", "MongoDB", "Redis", "SQLite"}
	var services []model.Service
	for i, env := range environments {
		if i >= 21 {
			break
		}
		svc := model.Service{
			ID:              uuid.New(),
			EnvironmentID:   env.ID,
			ServerID:        servers[rand.Intn(len(servers))].ID,
			Name:            fmt.Sprintf("%s-%s", apps[0].AppName, serviceTypes[i%5]),
			Type:            serviceTypes[i%5],
			URL:             fmt.Sprintf("https://%s.example.com", serviceTypes[i%5]),
			Repository:      fmt.Sprintf("https://github.com/ronaldocristover/%s", serviceTypes[i%5]),
			StackLanguage:   "Go",
			StackFramework:  frameworks[i%5],
			DBType:          dbTypes[i%5],
			DBHost:          servers[rand.Intn(len(servers))].IP,
		}
		db.Create(&svc)
		services = append(services, svc)
	}
	fmt.Printf("✅ Seeded %d services\n", len(services))

	// --- 6. Monitoring Configs (21, 1 per service) ---
	for _, svc := range services {
		enabled := rand.Intn(10) > 1 // 90% enabled
		interval := 30 + rand.Intn(5)*30 // 30, 60, 90, 120, 150
		cfg := model.MonitoringConfig{
			ID:                 uuid.New(),
			ServiceID:          svc.ID,
			Enabled:            enabled,
			PingIntervalSeconds: interval,
			TimeoutSeconds:     10,
			Retries:            3,
		}
		db.Create(&cfg)
	}
	fmt.Printf("✅ Seeded %d monitoring configs\n", len(services))

	// --- 7. Monitoring Logs (51) ---
	statuses := []string{"up", "down"}
	now := time.Now()
	for i := 0; i < 51; i++ {
		svc := services[rand.Intn(len(services))]
		status := statuses[0]
		statusCode := 200
		errMsg := ""
		if rand.Intn(10) == 0 { // 10% down
			status = "down"
			statusCode = 503
			errMsg = "connection timeout"
		} else {
			statusCode = []int{200, 201, 204, 301}[rand.Intn(4)]
		}
		log := model.MonitoringLog{
			ID:              uuid.New(),
			ServiceID:       svc.ID,
			Status:          status,
			ResponseTimeMs:  20 + rand.Intn(480),
			StatusCode:      statusCode,
			ErrorMessage:    errMsg,
			CheckedAt:       now.Add(-time.Duration(i*5) * time.Minute),
		}
		db.Create(&log)
	}
	fmt.Printf("✅ Seeded %d monitoring logs\n", 51)

	// --- 8. Deployments (13) ---
	methods := []string{"docker", "git-pull", "helm", "manual"}
	for i := 0; i < 13; i++ {
		svc := services[rand.Intn(len(services))]
		dep := model.Deployment{
			ID:            uuid.New(),
			ServiceID:     svc.ID,
			Method:        methods[i%4],
			ContainerName: fmt.Sprintf("%s-container", svc.Name),
			Port:          3000 + i,
			Config:        fmt.Sprintf(`{"replicas": %d, "auto_scale": true}`, 1+i%3),
		}
		db.Create(&dep)
	}
	fmt.Printf("✅ Seeded %d deployments\n", 13)

	// --- 9. Backups (7) ---
	schedules := []string{"0 2 * * *", "0 3 * * 0", "0 0 1 * *"}
	backupStatuses := []string{"success", "pending", "failed"}
	for i := 0; i < 7; i++ {
		svc := services[rand.Intn(len(services))]
		bk := model.Backup{
			ID:              uuid.New(),
			ServiceID:       svc.ID,
			Enabled:         true,
			Path:            fmt.Sprintf("/backups/%s/", svc.Name),
			Schedule:        schedules[i%3],
			LastBackupTime:  &[]time.Time{time.Now().Add(-time.Duration(i*12) * time.Hour)}[0],
			Status:          backupStatuses[i%3],
		}
		db.Create(&bk)
	}
	fmt.Printf("✅ Seeded %d backups\n", 7)

	fmt.Println("\n🎉 Seed complete!")
	fmt.Printf("  Users:              5\n")
	fmt.Printf("  Apps:               7\n")
	fmt.Printf("  Servers:            5\n")
	fmt.Printf("  Environments:       %d\n", len(environments))
	fmt.Printf("  Services:           %d\n", len(services))
	fmt.Printf("  Monitoring Configs: %d\n", len(services))
	fmt.Printf("  Monitoring Logs:    51\n")
	fmt.Printf("  Deployments:        13\n")
	fmt.Printf("  Backups:            7\n")
}
