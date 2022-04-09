FROM tinybirdco/tinybird-cli-docker 
RUN apt-get update \
 && apt-get install wget unzip zip -y