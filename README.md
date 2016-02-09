Minproject :: "Postback Delivery"
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

How to Install (Steps I follwed for Linux)
--------------

- Install Apache Server
apt-get install apache2-bin
- Install PHP
sudo apt-get install php5 libapache2-mod-php5 php5-mcrypt
- Install Go
apt-get install gccgo-go
- Install Make
apt-get install make
- Install Git
apt-get install git
- Install Predis library using Pear
apt-get install php-pear
pear channel-discover pear.nrk.io
pear install nrk/Predis
- Set Up Go's environment path
export GOPATH=$HOME/go
- Install Redigo library
go get github.com/garyburd/redigo/redis
- Install and setup a local Redis server
Followed [these steps](http://redis.io/topics/quickstart) including under "installing Redis more properly".

- Place "deliver.go" in the default directory according to your GOPATH
I placed mine at: $GOPATH/src/github.com/23caterpie/postback_delivery

- Place "ingest.php", "echoPost.php", and the "i" folder in your default server directory
I placed mine at: /var/www/html


How to Run
----------

... more to come