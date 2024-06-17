package command

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"scrapper/domain/entity"
	userPgx "scrapper/domain/repository/image/pgx"
	"scrapper/domain/service/image"
	"scrapper/infrastructure/godotenv"
	logger "scrapper/infrastructure/log"
	"scrapper/infrastructure/log/zerolog"
	pgxInfra "scrapper/infrastructure/pgx"
	imgDown "scrapper/utils/image"
	"strconv"
	"strings"
	"time"
)

func Boot() {
	logger := zerolog.NewLogger()
	env := godotenv.NewEnv()
	env.Load()

	conn, err := pgxInfra.SetupPool(env.DATABASE_HOST)
	if err != nil {
		logger.Error(err)
		panic(err)
	}
	sd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	sd = filepath.Join(sd, "images")

	imageRepo := userPgx.NewImageRepository(conn)
	addrService := image.NewService(logger, imageRepo, sd)

	method, proxy, err := getMethodAndProxyFromStdin()
	if err != nil {
		logger.Error(err)
		panic(err)
	}

	switch method {
	case "create":
		create(sd, addrService, proxy)
	case "read":
		read(addrService, logger)
	default:
		fmt.Println("Invalid method. Please enter 'create' or 'read'.")
	}
}

func getMethodAndProxyFromStdin() (string, bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Please enter the method (create|read): ")
	method, err := reader.ReadString('\n')
	if err != nil {
		return "", false, err
	}
	method = strings.TrimSpace(method)

	var proxy bool
	if method == "create" {
		fmt.Print("Do you want to use proxy? (yes|no) : ")
		proxyStr, err := reader.ReadString('\n')
		if err != nil {
			return "", false, err
		}
		proxyStr = strings.TrimSpace(proxyStr)
		proxy = (proxyStr == "yes")
	}

	return method, proxy, nil
}

func read(addrService *image.Service, logger logger.Logger) {
	for {
		count, err := getCountFromStdin()
		if err != nil {
			log.Fatalf("Error getting count from stdin: %v", err)
		}
		ch := make(chan *entity.Image, 50)
		startTime := time.Now()
		go addrService.Read(count, ch)
		for img := range ch {
			logger.Info(fmt.Sprintf("read image: %s", img.File))
		}

		elapsedTime := time.Since(startTime)
		fmt.Printf("Time taken: %s\n", elapsedTime)
	}
}

func create(sd string, addrService *image.Service, proxy bool) {
	for {
		count, err := getCountFromStdin()
		if err != nil {
			log.Fatalf("Error getting count from stdin: %v", err)
		}
		dr := imgDown.NewDownloadResizer(sd, count, zerolog.NewLogger(), proxy)

		startTime := time.Now()
		done := make(chan bool)
		addrService.Create(dr, done)
		<-done
		elapsedTime := time.Since(startTime)
		fmt.Printf("Time taken: %s\n", elapsedTime)
	}
}

func getCountFromStdin() (uint64, error) {
	fmt.Print("Please enter the images count you want: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	countStr := strings.TrimSpace(input)
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, errors.New("count number must be positive")
	}
	return uint64(count), nil
}
