package main

import (
"fmt"
"log"
"os"
"strings"

"github.com/go-resty/resty/v2"
tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
"github.com/joho/godotenv"
)

const DefaultMessage = "Welcome to the Temperature Bot!\n\nI can help you get the current temperature for any city.\n\nAvailable commands:\n/temperature [city] - Get the current temperature for a city (defaults to Cape Town)"

func main() {

	fmt.Println("Bot starting...")
	// Load environment variables from .env file if it exists
	// This allows storing sensitive data like BOT_TOKEN securely
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Get the bot token from environment variable
	// This is sensitive information, so it's stored in env vars for security
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable not set")
	}

	// Create a new bot instance using the token
	// The token is provided by BotFather when registering the bot
	bot, err := tgbotapi.NewBotAPI("8410367130:AAFosQHna9C2qrtOVF1lPWJQRQOGJVu9zSY")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	// Print confirmation that the bot is authorized
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Configure updates channel to receive messages
	// Timeout is set to 60 seconds, and we allow updates without offset
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Get the updates channel from the bot
	updates := bot.GetUpdatesChan(u)

	// Loop through each update received
	for update := range updates {
		// Skip updates that don't contain messages
		if update.Message == nil {
			continue
		}

		// Get the message text
		msg := update.Message

		// Check if the message starts with /start command
		if strings.HasPrefix(msg.Text, "/start") {
			reply := tgbotapi.NewMessage(msg.Chat.ID, DefaultMessage)
			bot.Send(reply)
		} else if strings.HasPrefix(msg.Text, "/temperature") {
			// Extract the city from the message, default to "Cape Town" if no city provided
			city := "Cape+Town"
			parts := strings.Fields(msg.Text)
			if len(parts) > 1 {
				city = strings.Join(parts[1:], "+")
			}

			// Fetch temperature data from weatherstack API
			temperature, err := getTemperature(city)
			if err != nil {
				log.Print(err)
				// Send error message if temperature fetch fails
				reply := tgbotapi.NewMessage(msg.Chat.ID, "Sorry, I couldn't fetch the temperature for "+city)
				bot.Send(reply)
				continue
			}

			// Send the temperature information back to the user
			reply := tgbotapi.NewMessage(msg.Chat.ID, temperature)
			bot.Send(reply)
		} else {
			// Send default message for unknown commands
			reply := tgbotapi.NewMessage(msg.Chat.ID, DefaultMessage)
			bot.Send(reply)
		}
	}
}

// TemperatureResponse represents the temperature data from weatherstack API
type TemperatureResponse struct {
	Current struct {
		Temperature int `json:"temperature"`
	} `json:"current"`
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
}

// getTemperature fetches temperature for a given city using weatherstack API
func getTemperature(city string) (string, error) {
	// Get API key from environment variable
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("WEATHER_API_KEY environment variable not set")
	}

	// Create a new resty client for making HTTP requests
	client := resty.New()

	// Construct the URL for weatherstack API
	url := fmt.Sprintf("http://api.weatherstack.com/current?access_key=%s&query=%s", apiKey, city)
	log.Print(url)
	// Make HTTP GET request to the API using resty
	resp, err := client.R().SetResult(&TemperatureResponse{}).Get(url)
	if err != nil {
		return "", err
	}

	// Check if the response status is OK
	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode())
	}

	// Parse the response
	result := resp.Result().(*TemperatureResponse)
	
	// Check if location data exists
	if result.Location.Name == "" {
		return "", fmt.Errorf("Could not find location data in API response")
	}

	// Return temperature string
	return fmt.Sprintf("Temperature in %s, %s: %d°C", result.Location.Name, result.Location.Country, result.Current.Temperature), nil
}
