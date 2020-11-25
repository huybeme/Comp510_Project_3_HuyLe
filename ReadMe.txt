Huy Le
COMP510 - Project 3

The intro screen contains the instructions to play the game but to reiterate briefly
    In the intro screen, enter a name to store into database and continue by pressing enter
    To move around, use the arrow keys to move in direction of the arrows
    Use A, S, D, or W, to shoot left, down, right, and up respectively.
    Hit every enemy sprite in order to move on.
        KhaiSprite (dragon) has two lives which will take two shots. Will not shoot.
        SophiaSprite (ninja) will Shoot but only has one life.
            The enemy sprite's ammo will disappear upon hitting any part of the maze but will not remove a player's life
    If players or enemies hit the wall, player will lose a live and enemies will disappear no matter number of their lives
        could not complete enemies from spawning over maze walls which will cause them to disappear immediately
    If player has lives <= -1 the game will end and it will prompt that you lose
    Clear all enemies to move onto next level - there are 3 levels to complete
    Enemies will move after every few seconds, didn't want to move too fast since hitting the maze will become more frequent
        enemies will move towards you
        if you collide into an enemy you will lose a life
        enemies shooting is more frequent than the movement

    Scoring scheme
    taking out SophiaSprite is 500 pts
    Taking out KhaiSprite's life is 200 pts, his second life will be an additional 300 pts on top
    if player dies is -100pts
    Score does update as you play

    Game's current level, player's name, number of lives remaining, and score is displayed

    At the end of the game, top 5 scores will show up
        BUT if your score is within the top five, it will not show on the list but it will prompt that there is a new high score
        and will show up on the top five list in the subsequent playing of the game

    Feel free to delete the database and run program. It should remake a new data base after program runs