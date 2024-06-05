import time
from gpiozero import PWMOutputDevice

# ESC settings
ESC_PIN = 18  # GPIO pin connected to the ESC signal wire
ESC_MIN_DUTY = 0.05  # Minimum duty cycle for the ESC
ESC_MAX_DUTY = 0.10 #0.1   # Maximum duty cycle for the ESC # => 0.125 (or 0.13) causes strong instablities, 1.2 is stable
ESC_RAMP_INCREMENT = 0.001 #0.001  # Increment for ramping up/down the speed
ESC_RAMP_DELAY = 0.01 #0.01  # Delay between each increment (in seconds)

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

#def accelerate_motor(target_duty):
#    current_duty = ESC_MIN_DUTY
#    while current_duty < target_duty:
#        set_esc_duty(current_duty)
#        time.sleep(ESC_RAMP_DELAY)
#        current_duty += ESC_RAMP_INCREMENT
#
#def decelerate_motor():
#    current_duty = esc.value
#    while current_duty > ESC_MIN_DUTY:
#        set_esc_duty(current_duty)
#        time.sleep(ESC_RAMP_DELAY)
#        current_duty -= ESC_RAMP_INCREMENT
#    set_esc_duty(ESC_MIN_DUTY)

def main():
    #calibrate_esc()
    #print("Waiting a second after calibration")
    #time.sleep(1)
    
    # CAUTION: Calibration Sould be run independently!!!
    print("setting esc duty to its min Value.")
    set_esc_duty(ESC_MIN_DUTY)
    print("Waiting for a second.")

    target_duty = 0.8 *  ESC_MAX_DUTY  # Adjust this value to set the desired motor speed

    print("Accelerating motor...")
    set_esc_duty(target_duty)
    # accelerate_motor(target_duty)
    print("Motor speed reached. Press Ctrl+C to stop.")
    try:
        while True:
            time.sleep(1)  # Maintain the motor speed
    except KeyboardInterrupt:
        print("Keyboard interrupt detected. Stopping motor...")
        set_esc_duty(ESC_MIN_DUTY)
 #       decelerate_motor()
    finally:
        esc.close()
        print("Motor stopped and PWM output closed.")

if __name__ == "__main__":
    main()