/*
Package bongo is an elegant static website generator. It is designed to be simple minimal
and easy to use.

Bongo comes in two flavors. The commandline applicaion and the library.

Commandline

The commandline application can be found in cmd/bongo directory

and you can install it via go get like this

	go get github.com/gernest/bongo/cmd/bongo

Or just download the latest binary here https://github.com/gernest/bongo/releases/latest

To build your project foo.

*  You can specify the path to foo

	bongo build --source path/to/foo

* You can run at  the root of foo

	cd path/to/foo

	bongo build

To serve your project locally. This will run a local server at port http://localhost:8000.
The project will be rebuilt if any markdown file changes.

*  You can specify the path to foo

	bongo serve --source path/to/foo

* You can run at  the root of foo

	cd path/to/foo

	bongo serve


The generated website will be in the directory _site at the root of your foo project.


The Website Project Structure

There is no restriction on how you arrange your project. If you have a project foo.
It will be somewhare in a directory named foo. You can see the example in testdata/sample directory.


Bongo only process markdown files found in your project root.Supported file extensions
for the markdown files are

	.md , .MD , .mdown, and .markdown

This means you can put your markdown files in any nested directories inside your project
and bongo will process them without any problem. Bongo support github flavored markdown

Optionaly, you can add sitewide configuration file `_bongo.yml` at the root of your project.
The configuration is in yaml format. And there are a few settings you can change.

	static
	  This is a list of static directories(relative from the project root). If defined
	  the directories will be copied to the output directory as is.

	title
	  The string representing the title of the project.

	subtitle
	  The string representing subtitle of the project

	theme
	  The name of the theme to use. Note that, bongo comes with a default theme called gh.
	  Only if you have a theme installed in the _themes directory at the root of your project
	  will you neeed to specify this.


Themes

There is  loose restrictions in the how to create your own theme. What matters is that you have
the following templates.

	index.html
		- used to render index pages for sections

	home.html
		- used to render home page

	page.html
		- used to render arbitrary pages

	post.html
		- used to render the posts


These templates can be used in project, by setting the view value of frontmatter. For instance
if I set view to post, then post.html will be used on that particular file.

IMPORTANT: All static contents should be placed in a diretory named static at the root of the
theme. They will be copied to the output directory unchanged.

All themes custom themes should live under the _theme directory at the project root. Please
see testdata/sample/_themes for an example.


Frontmatter

Bongo support frontmatter. And it is recomended every post(your markdown file) should have
a frontmatter. For convenience, only YAML frontmatter is supported by default. And you can
add it at the beginning of your file like this.

	---
	title: chapter one
	section: blog
	---

	Your post contents goes here.

Important frontmatter settings,

	title
		-The title of the post

	section
		- this acts as a category of sort. You can specify any section that the
		post will reside.

		Sections are in the form of relative directory paths. for instance the following
		are valid sections blog, blog/funny, blog/happy, blog/stuffs.

		If you specify section as blog/golang. bongo will put the generated html files in the
		folder named blog/golang. And you can referance your post by blog/golang/mypost.html.
		where mypost is the name of your markdown file.

		The default section is home.

	view
		- specifies the template to render the content.Defaults to post.



The Library

Bongo is modular, and uses interfaces to define its components.The most important interface is
the Generator interface.

So, you can implement your own Generator interface, and pass it to the bongo library to have your
own static website generator with your own rules.

I challenge you, to try implementing different Generators. Or, implement different components of the
generator interface. I have default implementations shipped with bongo.

*/
package bongo
