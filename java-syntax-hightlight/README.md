#Syntax Highlight
Simple project to create images from source files

Uses [picocli](https://picocli.info/) for command line parsing
and [RSyntaxArea](https://bobbylight.github.io/RSyntaxTextArea/) for higlighting

##

how to build

Java 17 or higher required

build:
./mvnw clean package assembly:single

target/SyntaxHighlight-1.0-SNAPSHOT-jar-with-dependencies.jar is an resulting artifact

##Usage

```
Usage: java -jar  SyntaxHighlight-1.0-SNAPSHOT-jar-with-dependencies.jar -i=INPUT [-l=LANG] -o=OUTPUT [-p=PADDINGS] [-t=THEME]
-i, --input=INPUT         input text file
-l, --lang=LANG           Language
-o, --output=OUTPUT       output image file
-p, --paddings=PADDINGS   padding around image
-t, --theme=THEME         theme to use
```


