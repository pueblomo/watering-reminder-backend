package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Plant struct {
  ID int `json:"id"`
  Alias  string `json:"alias"`
  Name string `json:"name"`
  Infos string `json:"infos"`
  LightBadge string `json:"lightBadge"`
  LightDescription string `json:"lightDescription"`
  WaterBadge string `json:"waterBadge"`
  WaterDescription string `json:"waterDescription"`
  MoistBadge string `json:"moistBadge"`
  MoistDescription string `json:"moistDescription"`
  ImageName string `json:"imageName"`
  UserName string `json:"userName"`
  Date time.Time `json:"date"`
}
var db *gorm.DB

func main(){
	connectToDatabase()
	app := fiber.New()
	app.Use(cors.New())
	setRoutes(app)
	log.Println("Running...")
	log.Fatalln(app.Listen(":8080"))
}

func connectToDatabase(){
	
	pw := viper.GetString("DBPASSWORD")
	url := viper.GetString("DBURL")
	dsn := fmt.Sprintf("host=%s user=postgres password=%s port=5432 sslmode=disable",url,pw)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err!=nil{
		log.Fatalln(err)
	}
	db.AutoMigrate(&Plant{})
}

func setRoutes(app *fiber.App){
	app.Static("/images","./images")

	plantGroup := app.Group("/plant")
	plantGroup.Get("/", func (c *fiber.Ctx) error {
		return getPlants(c)
	})

	plantGroup.Post("/", func (c *fiber.Ctx) error {
		return postPlants(c)
	})

	plantGroup.Put("/", func (c *fiber.Ctx)  error {
		return putPlants(c)
	})
}

func getPlants(ctx *fiber.Ctx) error{
	log.Println("Get Plants")
	user := getUserQuery(ctx)

	var plants []Plant
	result := db.Where("user_name = ?", user).Order("alias asc").Find(&plants)
	if result.Error != nil{
		log.Println(result.Error)
	}
	
	return ctx.Status(fiber.StatusOK).JSON(plants)
}

func postPlants(ctx *fiber.Ctx) error{
	log.Println("Post Plant")

	var newPlant Plant
	err := json.Unmarshal( []byte(ctx.FormValue("data")),&newPlant)
	if err != nil {
		log.Println(err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	newPlant.ImageName, err =saveImage(ctx, newPlant.UserName)
	if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	result := db.Create(&newPlant)
	if result.Error != nil{
		log.Println("Plant --> ", newPlant)
		log.Println(result.Error)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.SendStatus(fiber.StatusCreated)
}

func putPlants(ctx *fiber.Ctx) error {
	log.Println("Put Plant")

	var newPlant Plant
	err := json.Unmarshal( []byte(ctx.FormValue("data")),&newPlant)
	if err != nil {
		log.Println(err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	newImageName, err := saveImage(ctx, newPlant.UserName)
	if err != nil {
		log.Println(err)
	}else{
		newPlant.ImageName = newImageName
		var oldPlant Plant
		db.Where("user_name = ? and id = ?", newPlant.UserName, newPlant.ID).First(&oldPlant)
		deleteImage(oldPlant.ImageName)
	}

	result := db.Save(&newPlant)
	if result.Error != nil {
		log.Println("Plant --> ", newPlant)
		log.Println(result.Error)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.SendStatus(fiber.StatusCreated)
}

func saveImage(ctx *fiber.Ctx, name string) (string,error){
	log.Println("Save Image")

	file, err := ctx.FormFile("image")
	if err != nil {
		log.Println("Parse Error --> ", err)
		return "", err
	}

    uniqueId := uuid.New()
    filename := name + strings.Replace(uniqueId.String(), "-", "", -1)
	fileExt := strings.Split(file.Filename, ".")[1]

	image := fmt.Sprintf("%s.%s", filename,fileExt)

	err = ctx.SaveFile(file, fmt.Sprintf("./images/%s", image))
	if err != nil {
        log.Println("image save error --> ", err)
        return "", err
    }

	return image, err
}

func deleteImage(name string){
	log.Println("Delete Image")
	err := os.Remove(fmt.Sprintf("./images/%s", name))
	if err != nil {
		log.Println(err)
	}
}

func getUserQuery(ctx *fiber.Ctx) string {
	user := ctx.Query("user")
	if user == ""{
		log.Println("User Not Set")
		ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return user
}