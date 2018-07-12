# C2: A Compiler-Compiler
A meta-compiler written in Go that I created for fun.

# Specifics
It uses a language not too dissimilar to yacc/bison's to specify your language constructs.
Except that with C2, code blocks will consist of Go code instead of C++ code - obviously.
It functions currently on an LR0 parse table, which is subject to change.

# Structure
The code that should be used to parse is in the main directory,
whereas the code to modify when upgrading C2 is in 'nextGeneration'.
I've even included a handy bash script to copy the new generation out and replace the old one.
When making modifications to the language file (c2.c2l), C2 should be run on the file before copying out.

# Run
To parse a language file (a .c2l), run the binary with a '-p </path/to/lang.c2l>'.
To write your own language, you can generate default files by running the binary with a '-g </path/to/lang.c2l>'.
Fill out your language file (the .c2l) and modify your main.go as you please.
Don't modify 'parser.go', unless you like messing things up, or you're smarter than me.
Anyway, once you've filled out your language file, you can run C2 on your file.
If you have a syntax error, C2 will tell you (approximately (it's buggy sometimes)) which line and column it is at.
Additionally, if your grammar is not compatible with an LR0 parser, you will see some errors while parsing.
After the parse (with no errors), C2 will generate a 'yourlanguage.go' file that holds your symbols and parse table.
From here you can build and run your own parser. Horray...

# Other
I did make this for fun. I do not have plans of using it or improving it drastically at the moment.
I used yacc/bison as inspiration and a goal to work toward.
I researched LR0 (and a little bit of the other parse table types) on my own.
This was a very fun project. I was very excited when I could finally use it to build itself.
