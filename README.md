# Playlist-bot

Playlistbot is a music bot project for my own discord server.

Most of the technical decision were made because I wanted to learn the technologies that I used.

# Prerequisites
 - A kubernetes (1.27+) cluster
 - RabbitMQ
 - Redis
 - MySQL/MariaDB
 - An S3-compatible storage bucket




`config.yaml`
```yaml
http:
        address: '0.0.0.0:8081'
log:
        development: true
mysql:
        host: localhost
        user: root
        password: root
        dbname: playlist-bot
discord:
        botToken: sample-bot-token
youtube:
        implementation: "exec"
kubernetes:
        incluster: true
        namespace: playlist-bot
s3:
        endpoint: minio.minio.cluster.local:9000
        accessKey: 
        secretKey: 
        ssl: false

musicstore:
        bucket: playlists

amqp:
        username: guest
        password: guest
        address: localhost:5672
        tls: false
redis:
        address: 'localhost:6379'
        password: ''
        databaseIndex: 0

```

`discord-deployment.yaml`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: playlist-bot-discord
  namespace: playlist-bot
  labels:
    app.kubernetes.io/part-of: playlist-bot
    app.kubernetes.io/name: playlist-bot
    app.kubernetes.io/version: latest
    app.kubernetes.io/component: discord
    app.kubernetes.io/instance: discord
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/part-of: playlist-bot
      app.kubernetes.io/name: playlist-bot
      app.kubernetes.io/version: latest
      app.kubernetes.io/component: discord
      app.kubernetes.io/instance: discord
  template:
    metadata:
      labels:
        app.kubernetes.io/part-of: playlist-bot
        app.kubernetes.io/name: playlist-bot
        app.kubernetes.io/version: latest
        app.kubernetes.io/component: discord
        app.kubernetes.io/instance: discord
    spec:
      serviceAccount: playlist-bot
      volumes:
        - name: config
          configMap:
            name: playlist-bot
      containers:
        - name: playlist-bot
          image: mitaka8/playlist-bot:latest
          command:
          - "playlist-bot"
          - "discord"
          volumeMounts:
            - name: config
              readOnly: true
              mountPath: /etc/discordbot
          imagePullPolicy: Always
          tty: true
      restartPolicy: Always
```
