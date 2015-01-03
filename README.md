##streplace, maybe not the best name
st      = structured

replace = replace

Although joined together it looks like string replace.

But it is sort of that too.

The main use of this program is to injest structured data and paste out formatted data based on rules.

Much like a templating system.

There is a small grammer dialect that is definable in the program that can buildup collections from the data.

The collections can then be pushed through transformation logic to emit whatever. ATM I have mysql table/index structure scripts happening.

The data itself, and the grammer difinitions are tokenized by another library

The main itch I'm scratching here is to support script generation for mysql. So the grammer file is here along with an example.

Command line help:

Usage:
```
./streplace [cmt <string>] <gram file> [files ...] ... [<gram file> [files...]]
```

ex:

	./streplace gram ./mysql.gram ./example.tab	> example.tab.sql

the example files are also here.


###Short detour
github.com/chrhlnd/cmdlang

Its a light weight format that feels natural for me to type, and has the possiblity of being stream interpreted.

The rules are simple specify words seperated by white space. Carrige return is a delimiter unless its preceeded by a ','
then it is an extension of the previous line.

line comments are started with #

block comments are using ear muffs #( )#

'(' ')' pairs seperate sub commands/data. Its a way to have parent to child relationships.

"'" '"' provide quoting escape so you can have words with spaces or other quotes

```
# this is an example

this is data
this is more data
	,this is a continuation of the 2nd line
	,(this is a sub
		,command specified within the 2nd command
		another sub command note there is no ',' to lead after the CR)
```
