# EC2 Packet Limits

This is the code for a test we ran against AWS EC2 instances. The test floods the network with single byte packets to see if EC2 instances have a max number of packets.

### Prerequisites

**A server** We used AWS, but this code could run on any cloud and any server. [Sign up for AWS](https://aws.amazon.com)

**Golang** We used go1.10.3 on linux/amd64, but this code should run on most versions of Golang. [Download Golang](https://golang.org/dl/)

### Building the executable

Once you have golang installed, run the following:

```
cd <folder with packets.go>
go build packets.go
```

### Running the test

# Copy the executable `packets` or `packets.exe` to the test system using FTP, scp, or your preferred file transfer mechanism.
# Open a terminal on the test system
# Run the executable: `./packets`

By default, it will run for 48 hours, but can safely be killed using `Ctrl-c` at any time without loss of data. If you do plan to run this program for an extended period of time, use `tmux`, `screen`, or another tool to prevent the login session from dying and the process being killed prematurely.

### Results

The results of the test will be visible in three places:

* **Amazon Cloudwatch** Or the metric provider from your cloud of choice
* **stdout** Every second, statistics will be printed to the screen
* **file** About once an hour, and once when the process dies, the statistics will be saved to a gzipped JSON file in the same directory as the executable.

## Built With

* [Golang](https://golang.org/dl/)

## Authors

* **Matthew Barlocker** @ [Blue Matador](https://www.bluematador.com)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
