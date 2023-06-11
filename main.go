package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sort"

	_ "github.com/go-sql-driver/mysql"
	echo "github.com/labstack/echo/v4"

	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// DB connection
	var dbErr error
	dsn := "root:mrahmanpallab19977@tcp(localhost:3306)/pallab1?charset=utf8mb4&parseTime=True&loc=Local"
	db, dbErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if dbErr != nil {
		panic("falied to connect to db!")
	}

	// new instance of the app
	e := echo.New()
	fmt.Println(db)

	// routes
	e.GET("/", hello)
	e.GET("/activities", func(c echo.Context) error {
		return getActivities(c, db)
	})
	e.POST("/user", createNewUser)
	e.PATCH("/user", updateUser)
	e.DELETE("/user", deleteUser)

	// starting server
	err := e.Start(":8080")
	if err != nil {
		panic(err)
	}
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello World!")
}

type User struct {
	Id             uint   `json:"id" gorm:"primaryKey"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Country        string `json:"country"`
	ProfilePicture string `json:"profile_picture"`
}

type Activitiy struct {
	Id             uint   `json:"id"`
	FirstName      string `json:"first_name"`
	Country        string `json:"country"`
	ProfilePicture string `json:"profile_picture"`
	Points         uint   `json: "points"`
	Rank           int    `json:"rank"`
}

// returns an array of users containing points according to their activity
func getActivities(c echo.Context, db *gorm.DB) error {
	var useractivities []Activitiy
	resp := db.Raw("SELECT users.id, users.first_name, users.country, users.profile_picture, activities.points FROM ((users INNER JOIN activity_logs ON users.id = activity_logs.user_id) INNER JOIN activities ON activity_logs.activity_id = activities.id)").Find(&useractivities)
	if resp.Error != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}

	// calculatePoints() is used to calculate all the points of each user
	
	pointsMap := calculatePoints(useractivities)
	for i := range useractivities {
		useractivities[i].Points = pointsMap[useractivities[i].Id] // total points of a user is being assigned to the main array
	}

	// removeDuplicates() removes the duplicate elements of the responed array of db and maps each user's achieved point acccording to id
	uniqueUsers := removeDuplicates(useractivities)

	// Sorted useractivities based on points in descending order
	sort.SliceStable(uniqueUsers, func(i, j int) bool {
		return uniqueUsers[i].Points > uniqueUsers[j].Points
	})

	// raneked the sorted array
	rank := 1
	uniqueUsers[0].Rank = rank
	for i := 1; i < len(uniqueUsers); i++ {
		if uniqueUsers[i].Points < uniqueUsers[i-1].Points {
			rank++ // rank of each users are decided by comparing with previous user
		}
		uniqueUsers[i].Rank = rank
	}

	return c.JSON(http.StatusOK, uniqueUsers)
}

// calculatePoints() is used to calculate all the points of each user

func calculatePoints(users []Activitiy) map[uint]uint {
	pointsMap := make(map[uint]uint)
	for _, user := range users {
		pointsMap[user.Id] += user.Points // points are calculated for single user
	}
	return pointsMap
}

// removeDuplicates() removes the duplicate elements of the responed array of db and maps each user's achieved point acccording to id
func removeDuplicates(users []Activitiy) []Activitiy {
	uniqueUsers := make(map[uint]Activitiy)
	for _, user := range users {
		uniqueUsers[user.Id] = user // overwrites multiple instances with the updated one
	}

	result := make([]Activitiy, 0, len(uniqueUsers))
	for _, user := range uniqueUsers {
		result = append(result, user) // appends the user to the new slice result
	}
	return result
}

// creates a user of type User in db
func createNewUser(c echo.Context) error {
	reqBody := new(User)
	if err := c.Bind(reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	// create a new user object
	user := User{
		Id:             reqBody.Id,
		FirstName:      reqBody.FirstName,
		LastName:       reqBody.LastName,
		Country:        reqBody.Country,
		ProfilePicture: reqBody.ProfilePicture,
	}

	resp := db.Raw("INSERT INTO users (id, first_name, last_name, country, profile_picture) VALUES (?, ?, ?, ?, ?)", user.Id, user.FirstName, user.LastName, user.Country, user.ProfilePicture).Find(&user)
	if resp.Error != nil {
		// Return an error response if there's an issue with executing the SQL query
		return c.JSON(http.StatusInternalServerError, "Failed to create a new user")
	}
	return c.JSON(http.StatusOK, user)
}

// updates the user with specific id
func updateUser(c echo.Context) error {
	idStr := c.QueryParam("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user ID")
	}

	// Fetch the existing user from the database
	existingUser := User{}
	resp1 := db.Raw("SELECT * FROM users where id=?", id).Find(&existingUser)
	if resp1.Error != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	reqBody := new(User)
	if err := c.Bind(reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	// update the user fields if given, in the request (conditional updating)
	query := "UPDATE users SET"
	params := []interface{}{}
	if reqBody.FirstName != "" {
		query += " first_name = ?,"
		params = append(params, reqBody.FirstName)
	}
	if reqBody.LastName != "" {
		query += " last_name = ?,"
		params = append(params, reqBody.LastName)
	}
	if reqBody.Country != "" {
		query += " country = ?,"
		params = append(params, reqBody.Country)
	}
	if reqBody.ProfilePicture != "" {
		query += " profile_picture = ?,"
		params = append(params, reqBody.ProfilePicture)
	}

	// Removing last comma from query
	query = strings.TrimSuffix(query, ",")

	query += " WHERE id = ?"
	params = append(params, id)

	result := db.Exec(query, params...)
	err2 := result.Error
	fmt.Println(err2)

	return c.JSON(http.StatusOK, "Updated successfully")

}

// delete user from users and activity_log table for a specific id (general+bonus)
func deleteUser(c echo.Context) error {
	idStr := c.QueryParam("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid user ID")
	}

	query := `DELETE FROM users
              WHERE users.id = ?`

	resp := db.Exec(query, id)
	if resp.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User could not be deleted"})
	}

	// checking if deleting has occurred
	rowsAffected := resp.RowsAffected
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, "User deleted successfully")
}
