SET NAMES utf8mb4;

UPDATE user SET
  nickname = '球迷一号',
  email = 'fans01@tiyu.com',
  favorite_teams = '曼联,皇家马德里,巴塞罗那',
  favorite_players = '梅西,C罗,内马尔,哈兰德'
WHERE id = 4;
