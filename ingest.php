<?php
if(isset($_POST['"endpoint"']))
{
	echo "yup\n";
}
else
{
	echo "nada\n";
}

echo $HTTP_RAW_POST_DATA;

$json = json_decode(file_get_contents('php://input'), true);
echo $json;
?>