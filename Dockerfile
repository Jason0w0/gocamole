FROM alpine:latest

COPY gocamole /app/gocamole
COPY static/ /app/static/

EXPOSE 3000
CMD [ "/app/gocamole" ]
