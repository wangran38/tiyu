SET NAMES utf8mb4;

-- 1) 把现有「会员管理」从系统管理子菜单升级为顶级菜单（pid=0）
--    路径从 /system/member 改为 /member，component 改为 Layout（顶级菜单的容器）
UPDATE auth_rule
SET pid = 0,
    pathname = '/member',
    component = 'Layout',
    icon = 'peoples',
    title = '会员管理'
WHERE pathname = '/system/member';

-- 2) 新增「会员列表」子菜单（pid = 上一步升级后的会员管理 id）
INSERT INTO auth_rule (pid, type, icon, pathname, title, remark, ismenu, created, updated, weigh, status, component)
SELECT id, 'menu', 'user', '/member/list', '会员列表', '会员列表', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 100, '1', 'system/user/index'
FROM auth_rule WHERE pathname = '/member' AND pid = 0;

-- 3) 新增「会员流水」子菜单
INSERT INTO auth_rule (pid, type, icon, pathname, title, remark, ismenu, created, updated, weigh, status, component)
SELECT id, 'menu', 'chart', '/member/flow', '会员流水', '会员积分/等级变化流水', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 90, '1', 'system/userFlow/index'
FROM auth_rule WHERE pathname = '/member' AND pid = 0;
