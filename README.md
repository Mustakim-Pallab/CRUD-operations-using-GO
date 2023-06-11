# Backend Assignment 2

### GET:

https://localhost:8080/activities

- Joins the tables,
- Sums Based on total points accumulated by each user and rank them
- Returns an array of Object 

### POST :

https://localhost:8080/user

send from postman:
{
    "id" : 10000,
    "first_name" : "Mustakim",
    "last_name" : "Pallab",
    "country" : "Bangladesh",
    "profile_picture" : "mypropic"
}


### PATCH :

https://localhost:8080/user?id={id}

send from postman:
{
    "first_name" : "Pallab",
    "last_name" : "Mustakim",
    "country" : "Uganda",
    "profile_picture" : "mypropic"
}


### DELETE  :

https://localhost:8080/user?id={id}



