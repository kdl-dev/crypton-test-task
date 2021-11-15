# Client
***
## Instruction:

1. <b>Compile the application</b>
````
go build -o client cmd/main.go 
````

2. <b>Upload file</b>
````  
./client upload path_to_file your_password
````
* <b>Example</b>
````
./client upload /home/john/pictures/image.jpg MyImage123
````
3. <b>Unload file</b>
````
 ./client unload uid path_to_directory your_password
````
* <b>Example</b>
````
./client unload 1964721324 /home/john/pictures/ MyImage123
````
***
