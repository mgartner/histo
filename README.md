# histo

### How to Run

* Set the name of the cockroach binary `binaryName` in `main.go`, if it is not
  `cockroach`.
* Install dependencies: `go mod tidy`.
* Run the program: `go run main.go`.

### Example Output

```
Starting CockroachDB demo and executing SQL...
Column 'i' histogram contains 19996/20001 values in the ranges [1-10000], [100000-110000].
Column 's' histogram contains 19953/20001 values in the ranges [1-10000], [100000-110000].
Column 'f' histogram contains 19951/20001 values in the ranges [1-10000], [100000-110000].
```
