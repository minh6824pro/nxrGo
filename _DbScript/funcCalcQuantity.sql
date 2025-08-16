-- 1. Function tính available quantity cho một variant
DELIMITER $$

DROP FUNCTION IF EXISTS get_available_quantity$$

CREATE FUNCTION get_available_quantity(variant_id BIGINT)
    RETURNS INT
    READS SQL DATA
    DETERMINISTIC
BEGIN
    DECLARE original_qty INT DEFAULT 0;
    DECLARE reserved_qty INT DEFAULT 0;
    DECLARE available_qty INT DEFAULT 0;

    -- Lấy quantity gốc
SELECT COALESCE(quantity, 0) INTO original_qty
FROM product_variants
WHERE id = variant_id;

-- Tính tổng reserved quantity
SELECT COALESCE(SUM(oi.quantity), 0) INTO reserved_qty
FROM order_items oi
         INNER JOIN draft_orders do ON (
    (do.id = oi.order_id AND oi.order_type = 'draft_order' AND do.to_order IS NULL)
        OR
    (do.to_order = oi.order_id AND oi.order_type = 'order' AND do.to_order IS NOT NULL AND do.to_order != 0)
    )
WHERE oi.product_variant_id = variant_id;

-- Tính available quantity (không âm)
SET available_qty = GREATEST(original_qty - reserved_qty, 0);

RETURN available_qty;
END$$

DELIMITER ;

-- 2. Stored procedure để get batch available quantities (thay cho table function)
DELIMITER $$

DROP PROCEDURE IF EXISTS get_available_quantities_batch$$

CREATE PROCEDURE get_available_quantities_batch(IN variant_ids_str TEXT)
    READS SQL DATA
BEGIN
    -- Tạo temporary table để chứa results
    DROP TEMPORARY TABLE IF EXISTS temp_available_quantities;

    CREATE TEMPORARY TABLE temp_available_quantities (
        variant_id BIGINT,
        original_quantity INT,
        reserved_quantity INT,
        available_quantity INT
    );

    -- Insert results vào temp table
INSERT INTO temp_available_quantities (variant_id, original_quantity, reserved_quantity, available_quantity)
SELECT
    oq.variant_id,
    oq.original_quantity,
    COALESCE(rq.reserved_quantity, 0) as reserved_quantity,
    GREATEST(oq.original_quantity - COALESCE(rq.reserved_quantity, 0), 0) as available_quantity
FROM (
         -- Original quantities
         SELECT
             pv.id as variant_id,
             pv.quantity as original_quantity
         FROM product_variants pv
         WHERE FIND_IN_SET(pv.id, variant_ids_str) > 0
     ) oq
         LEFT JOIN (
    -- Reserved quantities
    SELECT
        oi.product_variant_id as variant_id,
        COALESCE(SUM(oi.quantity), 0) as reserved_quantity
    FROM order_items oi
             INNER JOIN draft_orders do ON (
        (do.id = oi.order_id AND oi.order_type = 'draft_order' AND do.to_order IS NULL)
            OR
        (do.to_order = oi.order_id AND oi.order_type = 'order' AND do.to_order IS NOT NULL AND do.to_order != 0)
        )
    WHERE FIND_IN_SET(oi.product_variant_id, variant_ids_str) > 0
    GROUP BY oi.product_variant_id
) rq ON rq.variant_id = oq.variant_id;

-- Return results
SELECT * FROM temp_available_quantities;

-- Cleanup
DROP TEMPORARY TABLE temp_available_quantities;
END$$

DELIMITER ;

