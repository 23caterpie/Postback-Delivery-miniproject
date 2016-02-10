Miniproject :: "Postback Delivery"
=================================

This webapp serves as a small scale simulation for data sychronization.

Software Stack
--------------

- [PHP](http://php.net/)
- [Redis](http://redis.io/)
- [Go](http://golang.org/)
- [Apache](https://httpd.apache.org/)

Libaries
--------

- [Redigo](https://github.com/garyburd/redigo/)
- [Predis](https://github.com/nrk/predis)

How to Install (Steps I use for Linux)
--------------

- Install Apache Server
	- apt-get install apache2-bin
- Install PHP
	- sudo apt-get install php5 libapache2-mod-php5 php5-mcrypt
- Install Go
	- apt-get install gccgo-go
- Install Make
	- apt-get install make
- Install Git
	- apt-get install git
- Install Predis library using Pear
	- apt-get install php-pear
	- pear channel-discover pear.nrk.io
	- pear install nrk/Predis
- Set Up Go's environment path
	- export GOPATH=$HOME/go
- Install Redigo library
	- go get github.com/garyburd/redigo/redis
- Install and setup a local Redis server
	- Follow [these steps](http://redis.io/topics/quickstart) including under "installing Redis more properly".
- Place "deliver.go" in the default directory according to your GOPATH
	- I placed mine at: $GOPATH/src/github.com/23caterpie/postback_delivery
- Place "ingest.php", "echoPost.php", and the "i" folder in your default server directory
	- I placed mine at: /var/www/html


How to Run
----------

- Start the Apache Server if it's not running already.
	- sudo service apache2 start
	- sudo service apache2 restart
- Redis should be running on local port 6379 already from the installation steps
- Change directory to where "deliver.go" is located and run it.
	- cd $GOPATH/src/github.com/23caterpie/postback_delivery
	- go run deliver.go
- The program will run until you close it with CTRL+C

Testing
-------

- Once you have everything running, you can test the project by sending the server a POST request to ingest.php
	- I used a Google Chrome App called [Postman](https://www.getpostman.com/)

-Send a POST request to http://\<server-ip\>/ingest.php or http://\<server-ip\>/i/ with a body of:
<pre><code>
{  
  "endpoint":{  
    "method":"post",
    "url":"http://localhost/echoPost.php/data?key={key}&value={value}&foo={bar}"
  },
  "data":[  
    {  
      "key":"Azureus",
      "value":"Dendrobates"
    },
    {  
      "key":"Phyllobates",
      "value":"Terribilis"
    }
  ]
}
</code></pre>
- You should receive a response code of 200 with a body "Success!\<br\>".
- Back on the server console, you should see log messages on standard output detailing delivering the data and receiving a response as well as the response code and body for both data objects.
- Since we set the endpoint url to localhost/echoPost.php, the response body will be a copy of the POST data it sent.
- The response bodies shoudld be:
	- {"bar":"","key":"Azureus","value":"Dendrobates"}
	- {"bar":"","key":"Phyllobates","value":"Terribilis"}
- By default the log file is "deliver.go.log" in the same directory as "deliver.go".
	- Check the log file for the same logs as outputted before.
	
Extra Configurations
-------------------

- To change the default unmatched key-value change "UNMATCHED_URL_KEY_VALUE" in deliver.go to the string you want unmatched keys in the url to be replaced with.
- To change the key for the Redis list used change both REDIS_LIST_NAME in deliver.go and $REDIS_LIST_NAME in ingest.php to the name the the list.
- To change what file logs are written to, change LOG_FILE_NAME in deliver.go to the name of that file.
- To see Trace level logs outputted to standard output, change SHOW_TRACES in delver.go to true.

