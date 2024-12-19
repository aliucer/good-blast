import boto3

dynamodb = boto3.resource('dynamodb', region_name='eu-north-1')
table = dynamodb.Table('TournamentEntries')

# Define query parameters
tournament_id = '2024-12-15'
user_id = '12ab9a21-3fc7-478e-b094-ba258a17951e'

print(f"Querying TournamentEntries for tournamentId: {tournament_id}, userId: {user_id}")

response = table.get_item(
    Key={
        'tournamentId': tournament_id,
        'userId': user_id
    }
)

item = response.get('Item')
if item:
    print("Item found:", item)
else:
    print("Item not found.")
