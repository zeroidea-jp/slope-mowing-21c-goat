import time
from gpiozero import PWMOutputDevice

# ESC settings
ESC_PIN = 18  # GPIO pin connected to the ESC signal wire
ESC_MIN_DUTY = 0.06  # Minimum duty cycle for the ESC
ESC_MAX_DUTY = 0.10 #0.1   # Maximum duty cycle for the ESC # => 0.125 (or 0.13) causes strong instablities, 1.2 is stable

# Initialize PWM output
esc = PWMOutputDevice(ESC_PIN, frequency=50, initial_value=0)

def set_esc_duty(duty_cycle):
    esc.value = duty_cycle

def calibrate_esc():
    print("Calibrating ESC...")
    set_esc_duty(ESC_MAX_DUTY)
    time.sleep(2)
    set_esc_duty(ESC_MIN_DUTY)
    time.sleep(2)
    print("ESC calibration complete.")


def main():
    calibrate_esc()
    print("Waiting a second after calibration")
    time.sleep(1)
    
if __name__ == "__main__":
    main()