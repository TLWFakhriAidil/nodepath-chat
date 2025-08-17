import re

# Change verbose logging from Info to Debug
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\sequence.go', 'r') as f:
    content = f.read()

# Replace specific log statements
replacements = [
    ('logrus.Infof("Retrieved %d steps for sequence', 'logrus.Debugf("Retrieved %d steps for sequence'),
    ('logrus.Infof("Step %d:', 'logrus.Debugf("Step %d:'),
    ('logrus.Infof("Processing sequence:', 'logrus.Debugf("Processing sequence:'),
    ('logrus.Infof("Found %d active sequences', 'logrus.Debugf("Found %d active sequences'),
    ('logrus.Infof("Creating %d steps for sequence', 'logrus.Debugf("Creating %d steps for sequence'),
]

for old, new in replacements:
    content = content.replace(old, new)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\sequence.go', 'w') as f:
    f.write(content)

print("Changed verbose sequence logging to Debug level")
