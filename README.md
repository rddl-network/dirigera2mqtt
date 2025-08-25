# dirigera2mqtt 

## API Usage

### POST /firmware/:mcu

Request body (JSON):
```json
{
	"ssid": "yourSSID",
	"pwd": "yourPassword"
}
```

Example curl command:
```sh
curl -X POST http://localhost:8080/firmware/esp32c6 \
	-H "Content-Type: application/json" \
	-d '{"ssid":"yourSSID","pwd":"yourPassword"}'
```
