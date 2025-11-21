# Log-montoring-for-go-gin-server (WIP)
This is a overall architecture for log monitoring in a golang gin server using promtail, loki and grafana dashboard

promtail is a software used to send log data from server to loki,
loki is another software from grafana which we can use to handle log files and send it to grafana dashboard

initially we can run the simple server golang server and ask it to store the logs in a volume so that the promtail can access the logs and send it to loki , loki can use this data and take it to grafana dashboard to give some insights

<img width="945" height="434" alt="image" src="https://github.com/user-attachments/assets/5ac0021a-644a-42ab-b2bd-02ab69bb4c6d" />


promtail has been deprected so that, we should use another software from grafana known as alloy
