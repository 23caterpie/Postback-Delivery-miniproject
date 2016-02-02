<?php

/*
 *name: Carter Currin
 *file: ingest.php
 *description: receives http requests with raw POST data in the following form:
 *  {  
 *    "endpoint":
 *    {  
 *      "method":"GET",
 *      "url":"http://sample-domain-endpoint.com/data?key={key}&value={value}&foo={bar}"
 *    },
 *    "data":
 *    [  
 *      {  
 *        "key":"Azureus",
 *        "value":"Dendrobates"
 *      },
 *      {  
 *        "key":"Phyllobates",
 *        "value":"Terribilis"
 *      }
 *    ]
 *  }
 *  
 *  for each "data" object in the accepted request, a "postback" object will be
 *  pushed to a local redis database with the endpoint and data information hash form:
 *  KEY           VALUES
 *  postback:###  "endpoint": {"method":"GET","url":"http:\/\/example-server.com\/data?key={key}&value={value}&foo={bar}"}
 *                "data": {"key":"Azureus","value":"Dendrobates"}
 */

require 'Predis/Autoloader.php';
Predis\Autoloader::register();

//echo $HTTP_RAW_POST_DATA."<br>";

//tries to decode JSON from the raw post data. If it can not, echos an error
if($content = json_decode(file_get_contents('php://input'), true))
{
	//var_dump($content);
	
	//varifies that the provided endpoint url is in valid form. Else echos an error
	if(filter_var($content['endpoint']['url'], FILTER_VALIDATE_URL))
	{
		//varifies the endpoint method is supported. Else echos an error
		if($content['endpoint']['method'] == GET ||
		$content['endpoint']['method'] == POST)
		{
			try
			{   //makes a coonection to the local redis database
				$redis = new Predis\Client();
				//makes separate "postback" objects for each received "data" object
				foreach($content['data'] as $data)
				{
					//redis keeps track of the newest postback object key number in "postback_tail"
					$id = $redis->incr("postback_tail");
					//postback keys are in the form "postback:####"
					$key = 'postback:'.$id;
					$redis->hset($key, 'endpoint', json_encode($content['endpoint']));
					$redis->hset($key, 'data', json_encode($data));
					
					echo 'key: '.$key.'<br>';
					echo '"endpoint": '.json_encode($content['endpoint']).'<br>';
					echo '"data": '.json_encode($data).'<br><br>';
				}
			}
			catch (Exception $e)
			{
				echo "Could not connect to Redis.<br>";
				echo $e->getMessage();
			}
		}
		else
		{
			echo 'Expected "method" to be GET or POST.<br>';
		}
	}
	else
	{
		echo 'Expected valid "url".<br>';
	}
}
else
{
	echo 'Raw POST Request not in JSON form.<br>';
}
?>