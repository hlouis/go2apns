#!/bin/sh

curl -d "token=a98c857e079bbe143a6a48a4e671b2a480af826de2ef3d9fb0172922fbd7b15f&expiration=0&priority=10&topic=mobi.xy3d.Go2ApnsTest&payload={\"aps\":{\"alert\":\"我是你爸爸\"}}" http://localhost:9090/push
