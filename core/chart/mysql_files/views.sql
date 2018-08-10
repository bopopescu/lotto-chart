CREATE OR REPLACE VIEW `v_buy_list` AS
	SELECT `a`.*,
				 `b`.`name`      AS `user_name`,
				 `b`.`comment`   AS `user_comment`,
				 `b`.`role_id`   AS `user_role_id`,
				 `b`.`role_name` AS `user_role_name`,
				 `c`.`name`      AS `lt_name`,
				 `c`.`name_cn`   AS `lt_name_cn`,
				 `c`.`enable`    AS `lt_enable`
	FROM `user_by_list` AS `a`
				 LEFT JOIN `users` AS `b` ON (`a`.`uid` = `b`.`id`)
				 LEFT JOIN `game_lts` AS `c` ON (`a`.`gid` = `c`.`gid`);