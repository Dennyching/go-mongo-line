# go-mongo-line


- Decription: use golang to implement golang line bot four function 
1. receive message (webhook) 
2. send message 
3. show user list
4. get message by user_id

- test step: 
1. clone project
2. add your own app.env and run "make local-debug"
3. run ngrok and add url "https://${your_ngrok_url}/api/webhook" to line Webhook settings
4. To test receive message, add your linebot id and send message to this bot
5. To test send message, call api "send message" provided in Postmen. Line bot should send message to your account.
6. To get userList, call api "get all user" provided in Postmen. Service should return json with all user_id recorded in system.
7. To get perticular user's message,call api "get message by user id" with your user_id as param in postman.Service should return json with all message recorded in system. 

- Postmen 
https://documenter.getpostman.com/view/1088678/2s935pr3wR 
