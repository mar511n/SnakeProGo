# SnakePro <img src="https://github.com/Mortim-Portim/SnakePro/blob/master/pics/logo.png" style="float:right" width="48" height="48" alt="Logo" /> <img src="https://github.com/Mortim-Portim/SnakePro/blob/master/pics/logo2.png" style="float:right" height="48" alt="Logo2" />


SnakeProGo is the Go version of SnakePro, a professional snake game with the following features:

    -- professional sounds and graphics
    -- items, such as bombs, lasers, speed ...
    -- player related statistics
    -- controller support
    -- support for android devices (used as controller)

## Installation:

    1. Download the zip file (Linux:SnakePro.zip / Win:SnakePro_Win.zip)
    2. Extract the files into a folder
    3. Extract res.zip into the same folder
    4. Run SnakePro / SnakePro.exe

## Parameter

    1. open res/params.txt
    2. change values & save
    3. restart game

## Creating a custom Texturepack

    duplicate "res/r_Standard" and rename the folder to "r_myRes"
    edit pictures and sounds or replace them with your own (but always keep the name)
    change the parameter "ResourceLoc" to "/r_myRes" (in "res/params.txt")

## Connect android phone
    
    to be able to connect an android phone and use it as a controller, you need to install SnakeControll.apk on your device.
    than connect both the PC and your phone to the same wifi and start SnakePro
    to connect the phone to the PC find out your computers IP-address by looking it up in your wifi-settings (it's also displayed in the console)
    type the IP and the Port(default=8080) into the app and click "connect" (for example: IP=192.168.83.31, Port=8080)
    
## License

SnakePro is licensed under Apache license version 2.0. See LICENSE file.
