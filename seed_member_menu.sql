SET NAMES utf8mb4;

INSERT INTO auth_rule (pid, type, icon, pathname, title, remark, ismenu, created, updated, weigh, status, component)
VALUES
(1, 'menu', 'peoples', '/system/member', '会员管理', '前台会员管理', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0, '1', 'system/user/index');
