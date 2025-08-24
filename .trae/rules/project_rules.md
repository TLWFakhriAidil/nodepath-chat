1. The port that used in this project is 8080 both local and Railways app.
2. The system need to handle high performance response and reply in real time.
3. The project will store all data into the MYSQL Database 5.7.
4. The deployment platform is Railways.
5. The project is open source and the code is hosted on GitHub.
6. The systems need to handle more than 3000+ real time device and user reply both sender and receiver at same time.
7. Each time finish fix, update, debug, create new any code function or anything update into the readme file.
8. The system naming for the database table will be end with the _nodepath.
9. The are no alter for the database table name because there may has same name, if ot same then create new database table but end it with _nodepath except for the database column for drop, create new column, update column.
10. The system default api url for AI is https://openrouter.ai/api/v1/chat/completions but if the id_device is SCHQ-S94 and SCHQ-S12 then api urls is https://api.openai.com/v1/chat/completions.
11. The model AI will be based on the device_setting_nodepath columns of api_key_option based on the id_device but if id_device is SCHQ-S94 and SCHQ-S12 then model AI will be used is gpt-4.1.
12. The AI payload will be used this on the whole system project $payload = [
        'model' => $model,
        'messages' => [
            ['role' => 'system', 'content' => $content],
            ['role' => 'assistant', 'content' => $lasttext],
            ['role' => 'user', 'content' => $currenttext]
        ],
        'temperature' => 0.67,  // Recommended setting
        'top_p' => 1,           // Keep responses within natural probability range
        'repetition_penalty' => 1, // Avoid repetitive responses
    ];.
13. The AI rules will be follow this $content = {AI PROMPT NODE DATA} "\n\n" . 
           "### Instructions:\n" . 
           "1. If the current stage is null or undefined, default to the first stage.\n" . 
           "2. Always analyze the user's input to determine the appropriate stage. If the input context is unclear, guide the user within the default stage context.\n" . 
           "3. Follow all rules and steps strictly. Do not skip or ignore any rules or instructions.\n\n" . 
           "4. **Do not repeat the same sentences or phrases that have been used in the recent conversation history.**\n" . 
           "5. If the input contains the phrase \"I want this section in add response format [onemessage]\":\n" . 
           "   - Add the `Jenis` field with the value `onemessage` at the item level for each text response.\n" . 
           "   - The `Jenis` field is only added to `text` types within the `Response` array.\n" . 
           "   - If the directive is not present, omit the `Jenis` field entirely.\n\n" . 
           "### Response Format:\n" . 
           "{\n" . 
           "  \"Stage\": \"[Stage]\",  // Specify the current stage explicitly.\n" . 
           "  \"Response\": [\n" . 
           "    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Provide the first response message here.\"},\n" . 
           "    {\"type\": \"image\", \"content\": \"https://example.com/image1.jpg\"},\n" . 
           "    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Provide the second response message here.\"}\n" . 
           "  ]\n" . 
           "}\n\n" . 
           "### Example Response:\n" . 
           "// If the directive is present\n" . 
           "{\n" . 
           "  \"Stage\": \"Problem Identification\",\n" . 
           "  \"Response\": [\n" . 
           "    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni.\"},\n" . 
           "    {\"type\": \"text\", \"Jenis\": \"onemessage\", \"content\": \"Kurang selera makan, sembelit, atau kerap demam?\"}\n" . 
           "  ]\n" . 
           "}\n\n" . 
           "// If the directive is NOT present\n" . 
           "{\n" . 
           "  \"Stage\": \"Problem Identification\",\n" . 
           "  \"Response\": [\n" . 
           "    {\"type\": \"text\", \"content\": \"Maaf kak, Layla kena reconfirm balik dulu masalah utama anak akak ni.\"},\n" . 
           "    {\"type\": \"text\", \"content\": \"Kurang selera makan, sembelit, atau kerap demam?\"}\n" . 
           "  ]\n" . 
           "}\n\n" . 
           "### Important Rules:\n" . 
           "1. **Include the `Stage` field in every response**:\n" . 
           "   - The `Stage` field must explicitly specify the current stage.\n" . 
           "   - If the stage is unclear or missing, default to first stage.\n\n" . 
           "2. **Use the Correct Response Format**:\n" . 
           "   - Divide long responses into multiple short \"text\" segments for better readability.\n" . 
           "   - Include all relevant images provided in the input, interspersed naturally with text responses.\n" . 
           "   - If multiple images are provided, create separate `image` entries for each.\n\n" . 
           "3. **Dynamic Field for [onemessage]**:\n" . 
           "   - If the input specifies \"I want this section in add response format [onemessage]\":\n" . 
           "      - Add `\"Jenis\": \"onemessage\"` to each `text` type in the `Response` array.\n" . 
           "   - If the directive is not present, omit the `Jenis` field entirely.\n" . 
           "   - Non-text types like `image` never include the `Jenis` field.\n\n";.
14. The system has three type personal user device command that is % for wablas provider, # for whacenter provider,this will be used when bot of AI reply not trigger to trigger to the receiver based on the current stage of the receiver and cmd for change status human into 1 that mean no ai reply and by default it 0 that mean ai reply is active.
15. The systems will used apikey $apikey from device_setting_nodepath columns of api_key but for the SCHQ-S94 and SCHQ-S12 will be used this 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.
16. The system will be record new prospect number only one time for each id_device mean that same phone number can be save multiple time but with same id_device.
17. The system will have a remote access into the databse using this mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway.
18. The system only use the MYSQL_URI to conencted to the database using the enviroment variables in railways for both system deploy in railways and local development.
19. The systems only used id_device for all process, so there will be no id_staff in the system.
20. The system database type input will not used the json type except the chatbot_flows_nodepath database.
21. The system database type output will not used the json type except the chatbot_flows_nodepath database.
22. The provider that used in this are Wablas and Whacenter only, this will get from the device_setting_nodepath columns of provider.
23. The system will used the instance from the device_setting_nodepath columns of instance for sending message.
24. for testing both localhost and railway used this id_device FakhriAidilTLW-001 and 60179645043 as the phone number.

