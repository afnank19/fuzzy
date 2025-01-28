### Fuzzy Finder

The program uses the concept of fuzzy search to search through the file system.

Contains two algorithms, namely 'Levenshtein Distance' and 'N-Grams' (Modified for this use case)
Levenshtein was too slow on large number of files so it was dropped, and the more performant N-Gram algorithm was used
Of course, further optimizations can definitely be made, but this code is from the time when I didn't know much Go.

The current code only searches file names, but the path can be added as well at a performance cost.
From my personal testing on a i5 1135G7, I was able to search to 200K to 300K files quite accurately.

The algorithm is a bit slower than what you'll find out there(such as fzf), but this still provides accurate results even if you misspell.

If you want to run it, you'll have to build it using the Go Compiler according to your OS.
