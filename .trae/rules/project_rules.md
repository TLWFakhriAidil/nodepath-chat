1. The port that used in this project is 8080 both local and Railways app.
2. The system need to handle high performance response and reply in real time.
3. The project will store all data into the MYSQL Database 5.7.
4. The deployment platform is Railways.
5. The project is open source and the code is hosted on GitHub.
6. The systems need to handle more than 3000+ real time device and user reply both sender and receiver at same time.
7. Each time finish fix, update, debug, create new any code function or anything update into the readme file.
8. The system naming for the database table will be end with the _nodepath.
9. The are no alter for the database table name because there may has same name, if ot same then create new database table but end it with _nodepath except for the database column for drop, create new column, update column.
10. The system default api url for AI is https://openrouter.ai/api/v1/chat/completions.
11. The model AI will be based on the device_setting_nodepath columns of api_key_option based on the id_device.
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
15. The systems will used this for ai generate reply, sanitize contain, satge, save data and etc:    $wa_no   = $whats->prospect_num;
            $wa_nama = $whats->prospect_nama;
        
            // Check time difference
            $jam          = $whats->balas;
            $currentTime  = now();
            $timeDifference = $currentTime->diffInSeconds($jam);
            if ($timeDifference < 4) {
                return;
            }
            
            $currenttext = $whats->conv_current ?? null;
        
            if (preg_match('/\bstage\s*:\s*(.+)/i', $currenttext, $matches)){
                $staged = trim($matches[1]);
                $whats->stage = $staged;
                $whats->save();
            }
        
            // 2) The user is already loaded
            $user = $whats->user;
            
            if (!$user) {
                return;
            }
        
                   
            // Keep the rest of your logic
            $stage      = $whats->stage;
            $instance   = $instance;
            $jenisapi   = $user->jenis_api;
            $apikey = $user->apiprovider;
            
            $apiurl = 'https://openrouter.ai/api/v1/chat/completions';
            $model  = $jenisapi;
        
            $lasttext    = $whats->conv_last ?? 'intro';
            
            if (empty($currenttext)) {
                return;
            }
        
            $behavePrompt  = $behave ? $behave->prompt : '';
            $closingPrompt = $closing ? $closing->prompt : '';
        
            $startTimez = microtime(true);
        
            $content = $ainodesprompt . "\n\n" . 
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
               "   - Non-text types like `image` never include the `Jenis` field.\n\n";
        
                $payload = [
                    'model' => $model,
                    'messages' => [
                        ['role' => 'system', 'content' => $content],
                        ['role' => 'assistant', 'content' => $lasttext],
                        ['role' => 'user', 'content' => $currenttext]
                    ],
                    'temperature' => 0.67,  // Recommended setting
                    'top_p' => 1,           // Keep responses within natural probability range
                    'repetition_penalty' => 1, // Avoid repetitive responses
                ];
     
            
            $client = new Client();
            
            try {
                $response = $client->post($apiurl, [
                    'headers' => [
                        'Authorization' => "Bearer {$apikey}",
                        'Content-Type'  => 'application/json',
                    ],
                    'json' => $payload,
                ]);
                $responseBody = json_decode($response->getBody(), true);
            }catch (\Exception $e) {
                Log::error('OpenRouter API error: ' . $e->getMessage());
                return;
            }
        
            $endTimez = microtime(true);
            $totalTimez = $endTimez - $startTimez;
            
            Log::info('Prompt' . $id_staff.': ' . $totalTimez . ' seconds');
          
            // The final response from the API
            if (!isset($responseBody['choices'][0]['message']['content'])) {
                Log::error("Invalid OpenRouter API response: " . json_encode($responseBody));
                return;
            }
    
            $replyContent = $responseBody['choices'][0]['message']['content'];
            $sanitizedContent = preg_replace('/^```json|```$/', '', trim($replyContent));
        
            // Attempt to decode JSON
            $data = json_decode($sanitizedContent, true);
            
            $last = null;
            $replyParts = [];
            
            if (is_array($data) && isset($data['Stage']) && isset($data['Response'])) {
                $stage      = $data['Stage'];
                $replyParts = $data['Response'];
            } elseif (preg_match('/Stage:\s*(.+?)\nResponse:\s*(\[.*?\])$/s', $replyContent, $matches)) {
                // Fallback for older format
                $stage       = trim($matches[1]);
                $responseJson = $matches[2];
                $replyParts  = json_decode($responseJson, true);
            } elseif (preg_match('/^\s*{\s*"Stage":\s*".+?",\s*"Response":\s*\[.*\]\s*}\s*$/s', $sanitizedContent)) {
                // Detect JSON
                $data = json_decode($sanitizedContent, true);
                if (is_array($data) && isset($data['Stage']) && isset($data['Response'])) {
                    $stage      = $data['Stage'];
                    $replyParts = $data['Response'];
                } else {
                    Log::error("Failed to parse specified JSON format: " . $sanitizedContent);
                    return;
                }
            } elseif (isset($replyParts[0]['content']) && preg_match('/^```json.*```$/s', $replyParts[0]['content'])) {
                // Encapsulated JSON within triple backticks
                $jsonContent    = preg_replace('/^```json|```$/', '', trim($replyParts[0]['content']));
                $decodedContent = json_decode($jsonContent, true);
        
                if (is_array($decodedContent) && isset($decodedContent['Stage']) && isset($decodedContent['Response'])) {
                    $stage      = $decodedContent['Stage'];
                    $replyParts = $decodedContent['Response'];
                } else {
                    Log::error("Failed to parse encapsulated JSON: " . $replyParts[0]['content']);
                    return;
                }
            } else {
                // Plain text fallback
                Log::warning("Plain text response detected. Defaulting to fallback handling.");
                $stage      = $stage ?? 'Problem Identification';
                $replyParts = [
                    ['type' => 'text', 'content' => trim($replyContent)]
                ];
            }
        
            if (!is_array($replyParts)) {
                Log::error("Failed to decode the response JSON properly.");
                return;
            }
        
            // Update stage in AIWhatsapp
            if(isset($stage)){
                $whats->stage = $stage;
            }
            
            if(isset($last)){
                $whats->conv_current = null;
                $whats->conv_last = $last;
                $whats->save();
            }
        
            // Time after we've parsed the AI response
            // EXACT same looping logic for messages
            $textParts           = [];
            $isOnemessageActive  = false;
            
            $idx = $whats->id_prospect;
        
            foreach ($replyParts as $index => $part) {
                if (!isset($part['type'], $part['content'])) {
                    Log::warning("Invalid response part structure: " . json_encode($part));
                    continue;
                }
        
                // if type=text and "Jenis"=onemessage
                if ($part['type'] === 'text' && isset($part['Jenis']) && $part['Jenis'] === 'onemessage') {
                    // Start collecting
                    $textParts[] = $part['content'];
                    $isOnemessageActive = true;
        
                    // If next part isn’t also onemessage, send combined
                    if (!isset($replyParts[$index + 1]) ||
                        !isset($replyParts[$index + 1]['Jenis']) ||
                        $replyParts[$index + 1]['Jenis'] !== 'onemessage') {
                            
                        $combinedMessage = implode("\n", $textParts);
                        $this->sendBhatMessage($wa_no, $combinedMessage, $instance, $delax);
                        
                        // Update conversation log
                        if(!isset($last)){
                            $newBotEntry = "BOT_COMBINED: " . json_encode($combinedMessage);
                            $whats->conv_last .= "\n" . $newBotEntry;
                            $whats->conv_current = null;
                            $whats->save();
                        }
                        // Reset temporary
                        $textParts          = [];
                        $isOnemessageActive = false;
                    }
                } else {
                    // If we just finished onemessage
                    if ($isOnemessageActive) {
                        $combinedMessage = implode("\n", $textParts);
                        $this->sendBhatMessage($wa_no, $combinedMessage, $instance,$delax);
                        
                        // Update log
                        if(!isset($last)){
                            $newBotEntry = "BOT_COMBINED: " . json_encode($combinedMessage);
                            $whats->conv_last .= "\n" . $newBotEntry;
                            $whats->conv_current = null;
                            $whats->save();
                        }
                        
                        $textParts          = [];
                        $isOnemessageActive = false;
                    }
        
                    // Now handle normal text or image
                    if ($part['type'] === 'text') {
                        $this->sendBhatMessage($wa_no, $part['content'], $instance,$delax);
                        
                        if(!isset($last)){
                            $newBotEntry = "BOT: " . json_encode($part['content']);
                            $whats->conv_last .= "\n" . $newBotEntry;
                            $whats->conv_current = null;
                            $whats->save();
                        }
                    } elseif ($part['type'] === 'image') {
                        $currentImageUrl = trim(urldecode($part['content']));
                        $this->sendChatMessage($wa_no, null, $currentImageUrl, $instance,$delax);
                        
                        $newBotEntry = "BOT: " . $currentImageUrl;
                        $whats->conv_last .= "\n" . $newBotEntry;
                        $whats->conv_current = null;
                        $whats->save();
                    }
                }
            }.
16. The system will be record new prospect number only one time for each id_device mean that same phone number can be save multiple time but with same id_device.
17. The system will have a remote access into the databse using this mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway.
18. The system only use the MYSQL_URI to conencted to the database using the enviroment variables in railways for both system deploy in railways and local development.
19. The systems only used id_device for all process, so there will be no id_staff in the system.
20. The system database type input will not used the json type except the chatbot_flows_nodepath database.
21. The system database type output will not used the json type except the chatbot_flows_nodepath database.
22. The provider that used in this are Wablas and Whacenter only, this will get from the device_setting_nodepath columns of provider.
23. The system will used the instance from the device_setting_nodepath columns of instance for sending message.
24. For testing flow will be using id_device FakhriAidilTLW-001, flow flow_ai_1756016272 and phone number 601137508067 for testing.
25. The system will always do the run test simulaion wihtout ncgo and fix any error occurs.

