# BLHeli_32_ESC_Driver
## Tested Environment
- Raspberry Pi 5, and its GPIO(18 and GND)
## How to use
### Installing packages
1. run `pip install -r requirements.txt`
### Calibration
*** CAUTION!! *** Motors could run at its MAXIMUM SPEED when we try to calibrate 
1. Connect one BLHeli_32_ESC, one brush less DC motor, and DC Power supply correctly, depending on the environment.
2. To calibrate, run: `python calibration_fr_bldc_control_steady.py`
3. Turn off the motor (DC power supply).
### Run motor
1. To run with sample code, run: `bldc_control_steady.py`
