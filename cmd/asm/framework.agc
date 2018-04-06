# Definitions of various registers.
ARUPT       EQUALS    10
QRUPT       EQUALS    12
TIME3       EQUALS    26
NEWJOB      EQUALS    67          # Location checked by the Night Watchman.

            SETLOC    4000        # The interrupt-vector table.

            # Come here at power-up or GOJAM
            INHINT                # Disable interrupts for a moment.
            # Set up the TIME3 interrupt, T3RUPT.  TIME3 is a 15-bit
            # register at address 026, which automatically increments every
            # 10 ms, and a T3RUPT interrupt occurs when the timer
            # overflows.  Thus if it is initially loaded with 037774,
            # and overflows when it hits 040000, then it will 
            # interrupt after 40 ms.
            CA        O37774
            TS        TIME3
            TCF       STARTUP    # Go to your "real" code.

            RESUME    # T6RUPT
            NOOP
            NOOP
            NOOP

            RESUME    # T5RUPT
            NOOP
            NOOP
            NOOP

            DXCH      ARUPT       # T3RUPT
            EXTEND                # Back up A, L, and Q registers
            QXCH      QRUPT
            TCF       T3RUPT

            RESUME    # T4RUPT
            NOOP
            NOOP
            NOOP

            RESUME    # KEYRUPT1
            NOOP
            NOOP
            NOOP

            RESUME    # KEYRUPT2
            NOOP
            NOOP
            NOOP

            RESUME    # UPRUPT
            NOOP
            NOOP
            NOOP

            RESUME    # DOWNRUPT
            NOOP
            NOOP
            NOOP

            RESUME    # RADAR RUPT
            NOOP
            NOOP
            NOOP

            RESUME    # RUPT10
            NOOP
            NOOP
            NOOP

# The interrupt-service routine for the TIME3 interrupt every 40 ms.  
T3RUPT      CAF     O37774      # Schedule another TIME3 interrupt in 40 ms.
            TS      TIME3

            # Tickle NEWJOB to keep Night Watchman GOJAMs from happening.
            # You normally would NOT do this kind of thing in an interrupt-service
            # routine, because it would actually prevent you from detecting 
            # true misbehavior in the main program.  If you're concerned about
            # that, just comment out the next instruction and instead sprinkle
            # your main code with "CS NEWJOB" instructions at strategic points.
            CS      NEWJOB

            # If you want to build in your own behavior, do it right here!

            # And resume the main program
            DXCH    ARUPT       # Restore A, L, and Q, and exit the interrupt
            EXTEND
            QXCH    QRUPT
            RESUME        

STARTUP     RELINT    # Reenable interrupts.

            # Do your own stuff here!

            # If you're all done, a nice but complex infinite loop that
            # won't trigger a TC TRAP GOJAM.
ALLDONE     CS      NEWJOB      # Tickle the Night Watchman
            TCF     ALLDONE

# Define any constants that are needed.
O37774      OCT     37774
