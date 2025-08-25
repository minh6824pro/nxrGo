DELIMITER $$

CREATE TRIGGER trg_update_order_done
AFTER UPDATE ON orders
FOR EACH ROW
BEGIN
    -- chỉ chạy khi status đổi sang 'DONE'
    IF NEW.status = 'DONE' AND OLD.status <> 'DONE' THEN

        -- cập nhật total_buy cho từng sản phẩm
        UPDATE products p
        JOIN product_variants pv ON pv.product_id = p.id
        JOIN order_items oi ON oi.product_variant_id = pv.id
        SET p.total_buy = p.total_buy + oi.quantity
        WHERE oi.order_id = NEW.id
          AND oi.order_type = 'order';

    END IF;
END$$

DELIMITER ;