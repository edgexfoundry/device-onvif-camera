
## Command Support
### Tapo C200 - User Management
Tapo returns `200 OK` for all User Management commands, but none of them actually
do anything. The only way to modify the users is through the Tapo app.

### Bosch - GetSnapshot
You must use `Digest Auth` or `Both` as the Auth-Mode in order for this to work.


