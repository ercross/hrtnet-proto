# This file exist to provide control for the program deployment on Heroku
# To run on a local file, make use of goals defined in the Makefile (can be found in project root directory)

# Apps on Heroku must listen on the port specified by the PORT environment variable
# $(echo PORT) is a linux command substitution used to collect the port for this program
# The web process specifies a command to build the binary && run the binary.
# To provide more flag argument for the run process, simply append the flag name and its value
# at the tail end of the web process value
web: go build -o bin/alpha-api ./cmd/*.go && ./bin/alpha-api -environment production -port $(echo PORT)