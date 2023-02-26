package controllers

import (
	"context"
	"fmt"
	"jwt/database"
	"jwt/models"
	"log"
	"net/http"
	"strconv"
	"time"

	helper "jwt/helpers"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client,"user")

var validate = validator.New()

func HashPass(password string) string{
	bytes,err := bcrypt.GenerateFromPassword([]byte(password),14)
	if err!= nil{
		log.Panic(err)
	}
	return string (bytes)
}

func VerifyPass(userPass string, providerPass string) (bool,string) {
	err:= bcrypt.CompareHashAndPassword([]byte(providerPass),[]byte(userPass))
	check := true
	msg := ""

	if err!= nil{
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}
	return check, msg
}

func Signup() gin.HandlerFunc{
	return func(ctx *gin.Context) {
	   c, cancle := context.WithTimeout(context.Background(),10*time.Second)
	   var user models.User

	   if err:= ctx.BindJSON(&user); err!= nil{
		ctx.JSON(http.StatusBadRequest,gin.H{"err":err.Error()})	
	   }
	   validationErr := validate.Struct(user)
	   if validationErr != nil {
		ctx.JSON(http.StatusBadRequest,gin.H{"err":validationErr.Error()})
	   }
	   password := HashPass(*&user.Password)
	   user.Password = password
	   emailCount, err := userCollection.CountDocuments(c, bson.M{"email":user.Email})
	   defer cancle()
	   if err!= nil {
		log.Panic(err)
		ctx.JSON(http.StatusBadRequest,gin.H{"err":"err!"})
	   }
	   phoneCount, err := userCollection.CountDocuments(c, bson.M{"phone":user.Phone})
	   defer cancle()
	   if err!= nil {
		log.Panic(err)
		ctx.JSON(http.StatusBadRequest,gin.H{"err":"!err"})
	   }
	   if emailCount>0 || phoneCount>0 {
		ctx.JSON(http.StatusInternalServerError,gin.H{"Err!":"this Email or password used"})
	   }
	   user.Created_at, _ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	   user.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	   user.ID = primitive.NewObjectID()
	   user.User_id = user.ID.Hex()
	   token, refreshToken,_ := helper.GenerateAllTokens(user.Email, user.First_name, user.Last_name, user.User_type, user.User_id)
	   user.Token = token
	   user.Refresh_token = refreshToken

	   res, err := userCollection.InsertOne(c, user)
	   if err != nil{
		msg := fmt.Sprintf("User Item Was Not Created")
		ctx.JSON(http.StatusInternalServerError,gin.H{"err":msg})
	}
	defer cancle()
	ctx.JSON(http.StatusCreated,res)

	}
}

func Login() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(),10*time.Second)
		var user models.User
		var foundUser models.User

		if err := ctx.BindJSON(&user);err!= nil{
			ctx.JSON(http.StatusBadRequest,gin.H{"err":err.Error()})
			
		}
		err := userCollection.FindOne(c, bson.M{"email":user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":"email or passwor is not correct"})
			
		}
		passwordIsValid, msg := VerifyPass(*&user.Password, *&foundUser.Password)
		defer cancel()
		if !passwordIsValid {
			ctx.JSON(http.StatusBadRequest,gin.H{"err":msg})
		}
		if foundUser.Email == ""{
			ctx.JSON(http.StatusInternalServerError,gin.H{"err":"user not found"})
		}
		token, refreshToken, _ := helper.GenerateAllTokens(foundUser.Email, foundUser.First_name, foundUser.Last_name, foundUser.User_type,foundUser.User_id)
		helper.UpdateAllToken(token, refreshToken, foundUser.User_id)
		err = userCollection.FindOne(c,bson.M{"user_id":foundUser.User_id}).Decode(&foundUser)
		if err != nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":err.Error()})
				return
		}
		ctx.JSON(http.StatusOK,foundUser)
	}
}

func GetUsers() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		if err:= helper.CheckUserType(ctx, "ADMIN"); err!=nil{
			ctx.JSON(http.StatusBadRequest,gin.H{"err":err.Error()})
			return
		}

		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err!= nil || recordPerPage<1{
			recordPerPage = 10
		}
		page, err:= strconv.Atoi(ctx.Query("page"))
		if err!= nil || page<1{
			page = 1
		}
		startIndex := (page - 1) * recordPerPage // e.g := page= 1 & recordPP = 10 => startIndex = 0 -> 9 
		startIndex,err = strconv.Atoi(ctx.Query("startIndex"))
		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}}}}
		res,err := userCollection.Aggregate(context.Background(),mongo.Pipeline{matchStage,groupStage,projectStage})
		// defer cancel()
		if err!= nil {
			ctx.JSON(http.StatusInternalServerError,gin.H{"err!":err.Error()})
		}
		var allUser []bson.M
		if err = res.All(context.Background(),&allUser);err!=nil{
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK,allUser[0])
	}
}

func GetUser() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")

		if err := helper.MatchUserToUid(ctx, userId); err!= nil{
			ctx.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		}
		c, cancel := context.WithTimeout(context.Background(),10*time.Second)
		defer cancel()

		var user models.User
		err:= userCollection.FindOne(c, bson.M{"user_id":userId}).Decode(&user)
		if err != nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{"err":err.Error()})
			return
		}
		ctx.JSON(http.StatusOK,user)
	}
}