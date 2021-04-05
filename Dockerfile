FROM alpine
COPY zwiftpower /
ENV SPREADSHEET=""
ENV LIMIT=0
ENV FILENAME=""
CMD ["/zwiftpower", "http", "-l", "3"]