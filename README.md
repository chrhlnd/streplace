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

Other databases or other things could also be scripted, currently MySQL is the main purpose of this project, I just made it generic.

## USE CASE: MySQL scripting
The main itch I'm scratching here is to support script generation for mysql. So the grammer file is here along with an example.

Command line help:

Usage:
```
./streplace [cmt <string>] <gram file> [files ...] ... [<gram file> [files...]]
```

ex:

	./streplace gram ./mysql.gram ./example.tab	> example.tab.sql

the example files are also here.

This ends up generating example.tab.sql that can just be piped into mysql and will upgrade/create tables in the 'crm' schema.

The scripts as they are now will only 'add' to the schema they won't ever delete anything I generally prefer this in practice if I need to delete
something I end up making a 1 off script and run it in a controlled manner.

The intented use is to create schema (structure / config data) structure. Then version control the struct.tab files and the config.tab files,
along with the grammer files and the tool binary. Then you can version control the structure of your db, and the config data just like normal
code files. To deploy you run the 'streplace' command on all your files with the appropriate grammers generating the .sql scripts, then pipe
them in to apply. You could also check in the scripts if you so desired, which might be appropriate if you're
versioning all release artifacts.

The project I used this for I wrote some stored procs to script out database tables/config data into the .tab format by inspecting the information_schema, there by having a full loop
for all the devs on the project. They can use mysql tools to create tables like normal then script them out as .tab files and check them into version
control or update existing ones. Then the build system just scripts everything in the .tab files and applies it to various environment databases to apply upgrades.

What does this all mean:

mysql.gram = convert 'table' structure files in to nice safe 'add/alter' scripts to apply to MySQL

mysql_data.gram = convert 'data' structure files trunc/insert statments to force apply 'config' data an application uses

```
Project repository:
	/Database/tables/
		schema.table1.tab
		schema.table2.tab
		schema.tableX.tab

	/Database/config/
		schema.<sys_config_data>.tab

```


```
Build system pulls
	repo:/Database/tables/*.tab
		pipe through 'streplace gram mysql.gram <file>.tab >> full_update.sql'
	repo:/Database/config/*.tab
		pipe through 'streplace gram mysql_data.gram <file>.tab >> full_update.sql'


```
... later apply updates to whatever deployment environment ..


```
	mysql <connection params> < full_update.sql
```
		

###Short detour
github.com/chrhlnd/cmdlang

Its a light weight format that feels natural for me to type, and has the possiblity of being stream interpreted.

The rules are simple specify words seperated by white space. Carrige return is a delimiter unless the next non whitespace token is ','
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
