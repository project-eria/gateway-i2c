# ERIA Project - Gateway for i2c

Tested on raspberry pi / raspbian

Supported devices:
- ads1115
- bmp388
- sht31d

## Configure
### Raspberry Pi OS
* Run sudo raspi-config.
* Use the down arrow to select 5 Interfacing Options
* Arrow down to P5 I2C.
* Select yes when it asks you to enable I2C
* Also select yes if it asks about automatically loading the kernel module.
* Use the right arrow to select the <Finish> button.
* Select yes when it asks to reboot.

```
sudo adduser eria spi
```
### DietPi OS
* Run dietpi-launcher
* Select "DietPi-Config"
* Select "4  : Advanced Options"
* Enable "I2C state"
* Select yes when it asks to reboot.

## Various
### i2c tools
```
sudo apt-get install -y i2c-tools
i2cdetect -y 1
```