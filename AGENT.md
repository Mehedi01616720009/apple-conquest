Game Flow - Apple Conquest

1. First Player Set His Name and Choose a Color and Select the Castle that he wants to play on.
2. Rest of the Castle is played by AI Players.
3. AI Players will play the game by following the rules of the game.
4. Everyone will develop their economy, build their army, upgrade army, castle, house and economic building.
5. Player can truce with AI Players. Declear war or rival with AI Players. Propose alliance with AI Players.
6. Player can send army to attack AI Players. And set garrison to defend AI Players.
7. Player can make vassals to AI Players.
8. Player can occupy castle after winning war. and manage the castle by himself.
9. AI Players randomly play activity as other game by following the rules and actions of the game.

Castles and AI Rulers Exists in the Game -

1. Cairo - Sultan Salahuddin Ayyubi
2. Alexandria - Emir Kazi Fadil
3. Kerak - Lord Reynald De Chatillon
4. Damascus - Sultan Nur Al Din Zengi
5. Jerusalem - King Baldwin IV
6. Tripoli - Lord Reymond III
7. Acre - Baron Balian of Ibelin
8. Edessa - Joscelin III
9. Aleppo - Emir Muzaffar Uddin Gokbori
10. Mosul - Emir Saifuddin Zengi

Rulers will be AI Players in the Game. Except Player select the castle to play on. Rest of the Castle will be played by AI Players. When Player vassalize the AI rulers of take the castles all. Then Player will be won. Player will lose if he have no castle.

Note - Game will be real time. Not a turn based game. Create also terminal interface for the game to play. Colorfull. Decorator will be hiphen -----. If possible can view a simple terminal map view of the game. That will show Occupied castles, lands and AI Players.

# New Issues

## UI Problems

### Action Buttons

- I do not see the build sawmill, market or any other buildings. It only shows farm and forge.
- Recruit only soldiers, not artillery button show.
- It shows attack from `CastleName` button. This button sends all field soldiers and artillery to attack. But it do not allow me to enter how many soldiers and artillery I want to send to attack.
- Now attack action happen immidiately. But I want to see the attack progress. Army Move one castle to another castle by 2 days. Attack stays for 1 day.

### Game Flow

- I see a Castle can attack only from its neighbor. It is correct. But Sometime the castle have no army and it's economy is not enough to recruit soldiers. So it needs to get army from other castle. But it cannot to send army from one castle to another castle of own realm.
- Also I can request army from Ally and Vassal. And I can send army to Ally and Vassal.
- [New Feature] I can trade with Ally and Vassal and in own realm.

### Functionality

- I see that days change every 5 seconds. But I want to see it change every 12 seconds.
- I see that AI Players all play to me. That means they all attack only me. But I want they also play with each other.
- I see my color is only different. That's the one of reason I know they play with me only. They should have own seperate color of different King. And I see expansion of them.
- I see that Truce, Alliance, Vassal offer have no logic. They accept instantly when I click the action. I describe the logics of these actions.

#### Alliance Action

- If I propose alliance. Then see that these logics:
    - My resource sum (gold, wood, food) should greater than that Player's resource sum. [Required]
    - Then compare army. Both army under 20% torallernce difference, then accept the alliance. [Required]
    - Example: I have 100 army and Player have 80 army OR I have 80 army and Player have 100 army. Accept 20% difference.

#### Vassal Action

- If I propose vassal. Then see that these logics:
    - My resource sum (gold, wood, food) should at least 60% greater than that Player's resource sum. [Required]
    - Then compare army. My army should at least 50% greater than that Player's army, then accept the vassal. [Required]
    - Example: I have 100 resources and Player should have <=40 resources. I have 100 army and Player should have <=50 army.

#### Truce Action

- If I propose vassal. Then see that these logics:
    - First Logic:
        - My resource sum (gold, wood, food) should at least 35-59% greater than that Player's resource sum. [Required]
        - Then compare army. My army should at least 40-59% greater than that Player's army, then accept the truce. [Required]
        - Then throw a probability of 50% to accept the truce.
        - Example: I have 100 resources and Player should have >40 and <=65 resources. I have 100 army and Player should have >40 and <=60 army. Probability of >50% to accept the truce.
    - Second Logic:
        - My resource sum (gold, wood, food) should at least 60% greater than that Player's resource sum. [Required]
        - Then compare army. My army should at least 60% greater than that Player's army, then accept the truce. [Required]
        - Example: I have 100 resources and Player should have <=40 resources. I have 100 army and Player should have <=40 army.
