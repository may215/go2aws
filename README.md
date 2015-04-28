[![GoDoc](https://godoc.org/github.com/may215/go2aws?status.svg)](https://godoc.org/github.com/may215/go2aws)

go2aws
===============

> [Go](http://golang.org) service securely fetch objects from Amazon s3 without using any load balancers/modules etc... 
> The motivation for this project was, first as always, love to experience the go language, second, we needed solution to fetch 
> very secured files from s3 with IAM rules, without counting on load balancers to reverse request to s3. 

Install
===============

    $ git clone https://github.com/may215/go2aws
    
Configuration
===============
You need to set the environment you want to use, e.g. (dev, qa, staging, production) 
You can change the configuration file(The main.go file contain explanations about each config element):

    "AwsKey": "Aws key string for amazon aws services"
    "awsSecret": "Aws secret string for amazon aws services"
	"awsRegion": "Amazon aws services location"
	"awsBasePath": "S3 base path"
	"awsBucketPath": "S3 file path in bucket"
	"awsEndPoint": "S3 end point domain"
	"bucketName": "The base s3 bucket"

Usage
===============

After you configure and run the service, you can browse it:

    $> go run go2aws
    $> curl http://localhost:8080/{format}/{fileName}

	The {format} replacement can be one of the following: json/xml/csv. (of course you can extend it).
	The {fileName} replacement can be the full name of the file as it appear on s3.


Examples:
===============

	curl -i http://localhost:8080/csv/file.csv

	curl -i http://localhost:8080/xml/file.xml

	curl -i http://localhost:8080/json/file.json

The JSON endpoint also supports JSONP, by adding a callback argument to the request query.

Example:
===============

	curl -i http://localhost:8080/json/8.8.8.8?callback=func

See http://en.wikipedia.org/wiki/JSONP for details on how JSONP works.

 * The package provides an http.Handler object that you can add to your HTTP server to fetch files from any s3 storage location. There's also an interface for crafting your own HTTP responses encoded in any format.
 * Check out the godoc reference.


Version
----

0.61

License
----

MIT

Author
----

Meir Shamay [@meir_shamay](https://twitter.com/meir_shamay)

**Free Software, Hell Yeah!**

[@meir_shamay]:https://www.twitter.com/meir_shamay
