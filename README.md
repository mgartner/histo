# histo

### How to Run

* Set the name of the cockroach binary `binaryName` in `main.go`, if it is not
  `cockroach`.
* Install dependencies: `go mod tidy`.
* Run the program: `go run main.go`.

### Example Output

```
Starting CockroachDB demo and executing SQL...
Column 'i' histogram matches 20000/20001 values in the ranges [1-10000], [100000-110000].
Column 's' histogram matches 19935/20001 values in the ranges [1-10000], [100000-110000].
```
