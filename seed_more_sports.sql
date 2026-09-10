SET NAMES utf8mb4;

INSERT INTO sports_category (name, icon, remark, weigh, status, rules, created, updated) VALUES
('棒球', '⚾', '棒球运动分类', 3, 'normal',
'<h2>棒球比赛规则</h2><p><strong>1. 比赛场地</strong>：扇形场地，内场呈菱形，边长 27.43 米，外场为弧形区域。</p><p><strong>2. 比赛时间</strong>：九局制，每局分上下半局，无固定时长。</p><p><strong>3. 队员人数</strong>：每队 9 人上场，投手、捕手、内野手、外野手各司其职。</p><p><strong>4. 得分</strong>：击球员击球后依次跑过一、二、三垒回到本垒得 1 分。</p><p><strong>5. 出局方式</strong>：三振出局、接杀、封杀、触杀等。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('网球', '🎾', '网球运动分类', 4, 'normal',
'<h2>网球比赛规则</h2><p><strong>1. 比赛场地</strong>：长方形场地，单打 23.77×8.23 米，双打宽 10.97 米，球网高度 0.914 米。</p><p><strong>2. 比赛时间</strong>：无固定时间，采用盘数制，先胜 6 局为 1 盘，三盘两胜或五盘三胜。</p><p><strong>3. 队员人数</strong>：单打 2 人，双打 4 人。</p><p><strong>4. 计分</strong>：每局分 15、30、40、胜，每盘先胜 6 局，平局时需净胜 2 局。</p><p><strong>5. 发球规则</strong>：每局轮流发球，发球需落在对角线发球区内，两次发球机会。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('田径', '🏃', '田径运动分类', 5, 'normal',
'<h2>田径比赛规则</h2><p><strong>1. 比赛场地</strong>：标准 400 米田径场，跑道 8 条，跳高、跳远、铅球等专项场地环绕四周。</p><p><strong>2. 比赛项目</strong>：径赛（短跑、中长跑、跨栏、接力）、田赛（跳高、跳远、三级跳、撑杆跳、铅球、标枪、铁饼、链球）。</p><p><strong>3. 队员人数</strong>：每项参赛选手 8-12 人，接力每队 4 人。</p><p><strong>4. 计时与判定</strong>：径赛按时间判定名次，田赛按距离或高度判定。</p><p><strong>5. 起跑规则</strong>：采用蹲踞式起跑，听枪声起跑，二次抢跑取消资格。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('乒乓球', '🏓', '乒乓球运动分类', 6, 'normal',
'<h2>乒乓球比赛规则</h2><p><strong>1. 比赛场地</strong>：长方形球台 2.74×1.525 米，高 0.76 米，球网高度 15.25 厘米。</p><p><strong>2. 比赛时间</strong>：无固定时间，每局 11 分，五局三胜或七局四胜制。</p><p><strong>3. 队员人数</strong>：单打 2 人，双打 4 人。</p><p><strong>4. 计分</strong>：每球得分制，先得 11 分且净胜 2 分胜一局，10 平后连得 2 分胜。</p><p><strong>5. 发球规则</strong>：每 2 分轮换发球，10 平后每 1 分轮换，发球需抛起 16 厘米以上。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('羽毛球', '🏸', '羽毛球运动分类', 7, 'normal',
'<h2>羽毛球比赛规则</h2><p><strong>1. 比赛场地</strong>：长方形场地 13.4×6.1 米（双打）/ 5.18 米（单打），球网高度 1.55 米。</p><p><strong>2. 比赛时间</strong>：无固定时间，每局 21 分，三局两胜制。</p><p><strong>3. 队员人数</strong>：单打 2 人，双打 4 人。</p><p><strong>4. 计分</strong>：每球得分制，先得 21 分且净胜 2 分胜一局，决胜局先得 11 分交换场地。</p><p><strong>5. 发球规则</strong>：发球需低于腰部，拍头朝下，对角线发球区。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('亲子活动', '👨‍👩‍👧', '亲子活动分类', 8, 'normal',
'<h2>亲子活动规则</h2><p><strong>1. 活动场地</strong>：户外草坪、运动场或室内亲子馆。</p><p><strong>2. 活动时间</strong>：通常 1-2 小时，含多个互动环节。</p><p><strong>3. 参与人数</strong>：以家庭为单位，每组成人 1-2 人，儿童 1-2 人。</p><p><strong>4. 项目类型</strong>：趣味接力、二人三足、亲子跳绳、合作搬运等。</p><p><strong>5. 评分原则</strong>：以完成时间、合作度、趣味性为评判标准，弱化竞技性。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('登山', '⛰️', '登山运动分类', 9, 'normal',
'<h2>登山活动规则</h2><p><strong>1. 活动场地</strong>：户外山地、自然景区或人造攀岩墙。</p><p><strong>2. 活动时间</strong>：单日往返或数日穿越。</p><p><strong>3. 参与人数</strong>：建议 4-12 人结组，配 1-2 名向导。</p><p><strong>4. 装备要求</strong>：登山鞋、登山杖、安全头盔、保暖衣物、急救包、充足饮水与食物。</p><p><strong>5. 安全原则</strong>：遵循"三人结组"原则，恶劣天气立即下撤，不得单独行动。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('露营', '🏕️', '露营活动分类', 10, 'normal',
'<h2>露营活动规则</h2><p><strong>1. 活动场地</strong>：营地、郊野公园或野外帐篷区。</p><p><strong>2. 活动时间</strong>：1-3 天为常见时长。</p><p><strong>3. 参与人数</strong>：每帐 2-6 人，营地总人数不限。</p><p><strong>4. 装备要求</strong>：帐篷、睡袋、防潮垫、炊具、照明、防蚊用品、应急药品。</p><p><strong>5. 营地规则</strong>：避开低洼地与河道，远离野生动物觅食区，垃圾全部带离。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('骑行', '🚴', '骑行运动分类', 11, 'normal',
'<h2>骑行比赛规则</h2><p><strong>1. 比赛场地</strong>：公路、山地、场地赛道或城市街道。</p><p><strong>2. 比赛时间</strong>：公路赛数小时，场地赛数分钟，绕圈赛 30-90 分钟。</p><p><strong>3. 队员人数</strong>：个人或团队计时赛，公路大组赛每队 4-7 人。</p><p><strong>4. 装备要求</strong>：合规自行车、头盔、骑行服、前后灯。</p><p><strong>5. 安全规则</strong>：右侧通行，不得借助机动车牵引，转弯前打手势示意后车。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('橄榄球', '🏈', '橄榄球运动分类', 12, 'normal',
'<h2>橄榄球比赛规则</h2><p><strong>1. 比赛场地</strong>：长方形场地 100×70 米，端区 10-22 米深。</p><p><strong>2. 比赛时间</strong>：四节 15 分钟（美式）或两个 40 分钟半场（英式）。</p><p><strong>3. 队员人数</strong>：美式 11 人，英式 15 人或 7 人制。</p><p><strong>4. 得分</strong>：达阵 6 分（美式）/5 分（英式），追加射门 1-2 分，射门 3 分。</p><p><strong>5. 传球规则</strong>：美式可向前传球，英式只能向后或横向传球。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('体操', '🤸', '体操运动分类', 13, 'normal',
'<h2>体操比赛规则</h2><p><strong>1. 比赛场地</strong>：12×12 米自由操场地，单项器械（鞍马、吊环、跳马、双杠、单杠、平衡木、高低杠）。</p><p><strong>2. 比赛时间</strong>：无固定时长，按项目分项进行。</p><p><strong>3. 队员人数</strong>：个人项目或团体赛 5 人一组。</p><p><strong>4. 评分</strong>：D 分（难度分）+ E 分（完成分）减扣分项，满分约 16 分。</p><p><strong>5. 比赛类型</strong>：资格赛、个人全能、单项决赛、团体决赛。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('攀岩', '🧗', '攀岩运动分类', 14, 'normal',
'<h2>攀岩比赛规则</h2><p><strong>1. 比赛场地</strong>：室内攀岩墙或天然岩壁，含抱石、难度、速度三种赛道。</p><p><strong>2. 比赛时间</strong>：速度赛数十秒，难度赛限时 6-8 分钟，抱石无时长。</p><p><strong>3. 队员人数</strong>：个人赛或团体接力赛。</p><p><strong>4. 装备要求</strong>：攀岩鞋、镁粉、安全带、动力绳、保护器。</p><p><strong>5. 评分</strong>：难度赛按攀爬高度排名，抱石按完成线路数，速度赛按登顶时间。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('游泳', '🏊', '游泳运动分类', 15, 'normal',
'<h2>游泳比赛规则</h2><p><strong>1. 比赛场地</strong>：标准 50 米或 25 米游泳池，8 条泳道，每道宽 2.5 米。</p><p><strong>2. 比赛时间</strong>：短距离数十秒，长距离 15 分钟以上。</p><p><strong>3. 队员人数</strong>：每项 8 名选手同场，接力每队 4 人。</p><p><strong>4. 泳姿</strong>：自由泳、仰泳、蛙泳、蝶泳、个人混合泳、接力。</p><p><strong>5. 出发规则</strong>：听到信号出发，抢跳取消资格；蛙泳与蝶泳需双手同时触壁。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('极限运动', '🏄', '极限运动分类', 16, 'normal',
'<h2>极限运动规则</h2><p><strong>1. 比赛场地</strong>：冲浪海域、滑板公园、BMX 赛道、翼装飞行空域。</p><p><strong>2. 比赛时间</strong>：每轮 30-60 分钟，多轮取最高分。</p><p><strong>3. 参与人数</strong>：个人或小组赛。</p><p><strong>4. 评分</strong>：难度分 + 完成度 + 创意分 + 速度分，综合排名。</p><p><strong>5. 安全要求</strong>：必须穿戴防护装备，恶劣天气禁止出海或飞行。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('长跑', '🏃‍♂️', '长跑运动分类', 17, 'normal',
'<h2>长跑比赛规则</h2><p><strong>1. 比赛场地</strong>：田径场（5000 米/10000 米）、公路（半马、全马、超马）。</p><p><strong>2. 比赛距离</strong>：5000 米、10000 米、半程马拉松 21.0975 公里、全程马拉松 42.195 公里。</p><p><strong>3. 队员人数</strong>：不限人数，大众组可数千人。</p><p><strong>4. 计时</strong>：枪声时或净时（芯片计时），取最佳成绩。</p><p><strong>5. 补给与关门</strong>：每 5 公里设补给点，全程马拉松通常设 6 小时关门时间。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

('台球', '🎱', '台球运动分类', 18, 'normal',
'<h2>台球比赛规则</h2><p><strong>1. 比赛场地</strong>：标准球台 9 英尺（2.7 米），分为中式八球、九球、斯诺克三种主流玩法。</p><p><strong>2. 比赛时间</strong>：无固定时间，按局数或得分制判定胜负。</p><p><strong>3. 队员人数</strong>：1 对 1 单打或 2 对 2 双打。</p><p><strong>4. 计分</strong>：八球先打进己方全色或花色再打 8 号黑球；九球按编号顺序击打；斯诺克按红球彩球交替得分。</p><p><strong>5. 击球规则</strong>：母球先击中己方目标球，不得落袋母球或先击对方球，犯规判罚分。</p>', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
