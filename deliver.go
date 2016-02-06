package main

import "fmt"
import "github.com/garyburd/redigo/redis"

func main() {
	c, err := redis.Dial("tcp", ":6379")
	
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	defer c.Close()
	
	if(!redis.Bool(c.Do("EXISTS", "postback_head"))){
			c.Do("SET", "postback_head", 0)
	}
	
	for{
		postback_head := redis.Int(c.Do("GET", "postback_head"))
		c.Do("GET", "postback_head")
	}

	//set
	//c.Do("SET", "message1", "Hello World")

	//get
	//world, err := redis.String(c.Do("GET", "message1"))
	
	//if err != nil {
	//	fmt.Println("key not found")
	//}

	fmt.Println(world)
}
