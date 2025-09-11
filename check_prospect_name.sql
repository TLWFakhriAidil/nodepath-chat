-- Check if prospect_name was saved correctly
SELECT 
    id_prospect,
    id_device,
    prospect_num,
    prospect_name,
    stage,
    niche,
    created_at,
    updated_at
FROM ai_whatsapp_nodepath 
WHERE id_device = 'FakhriAidilTLW-001'
ORDER BY id_prospect DESC 
LIMIT 5;

-- Check the latest record specifically
SELECT 
    'Latest Record' as info,
    id_prospect,
    id_device,
    prospect_num,
    prospect_name,
    stage,
    conv_last,
    created_at
FROM ai_whatsapp_nodepath 
WHERE id_device = 'FakhriAidilTLW-001'
ORDER BY id_prospect DESC 
LIMIT 1;