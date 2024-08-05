#build
FROM eclipse-temurin:17-jdk-jammy AS build
COPY src /home/app/src
COPY pom.xml /home/app
COPY mvnw /home/app
COPY .mvn /home/app/.mvn
WORKDIR /home/app/
RUN ./mvnw clean package assembly:single

#package
FROM eclipse-temurin:17-jdk-jammy
WORKDIR /usr/local/lib
RUN useradd -ms /bin/bash appuser
RUN mkdir -p /images
COPY --from=build /home/app/target/SyntaxHighlight-1.0-SNAPSHOT-jar-with-dependencies.jar ./highlight.jar
USER appuser
ENTRYPOINT ["java", "-jar", "/usr/local/lib/highlight.jar"]
CMD ["--lang=java", "--input=/images/input.java", "--output=/images/output.png", "-p=10"]