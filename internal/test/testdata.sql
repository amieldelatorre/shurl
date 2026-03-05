DO $$
BEGIN
-- ---------------------------------------------------------------------------------------------------------
-- There should be 5 rows
-- add users
    INSERT INTO public.shurl_users VALUES 
        -- passwords is password
        ('019cb76d-23a3-7d94-9187-a702cbe03b3f', 'test1', 'test1@example.invalid', '$argon2id$v=19$m=262144,t=4,p=2$zCeK/qqV7BzHmKyAKKnE1g$mZ3YJMYRn/a+evEw6L9btgKMlrYxZUn+DhXt0DBoBWU', NOW(), NOW()),
        -- passwords is password
        ('019cb76d-645d-70a4-8c52-68dfa99cdfc6', 'test2', 'test2@example.invalid', '$argon2id$v=19$m=262144,t=4,p=2$Fe5VjdIgGxO6tran6jmeZw$zBwnwTOuntNWwgMzJvKZolPGkAsLBmAAaNrXc7BIBP8', NOW(), NOW()),
        -- passwords is password
        ('019cb76d-9c19-726d-ad64-7bc070a95940', 'test3', 'test3@example.invalid', '$argon2id$v=19$m=262144,t=4,p=2$pnBJNPrM3LSQiuEtycjukA$iHBkNO7q49pOHooynpRo5SagU0JTgafYAjcqr1SG7/8', NOW(), NOW()),
        -- passwords is password
        ('019cb76d-dada-73b5-81a7-6f132a876b2a', 'test4', 'test4@example.invalid', '$argon2id$v=19$m=262144,t=4,p=2$f5fyT93MmavHSCPOey27Vw$Hi0Z/349GHZvT+uU382wuvnzE+8QLyJV+9/k7EGNitA', NOW(), NOW()),
        -- passwords is password123
        ('019cbcdb-aaf4-7680-a3f7-8acef63e0151', 'test5', 'test5@example.invalid', '$argon2id$v=19$m=262144,t=4,p=2$5QPTScCJ0BaXlJrbuMXxjw$NalLgOx1YcKu62xCcG3eBkgic2KF60K2kR3bXcL3PqE', '2026-03-05 07:17:18.708701+00', '2026-03-05 07:17:18.708701+00');
-- ---------------------------------------------------------------------------------------------------------
    -- There should be 603 rows 
    -- add expired idempotency keys ----------------------------------------------------------------------
    INSERT INTO idempotency_keys (
        id,
        i_key,
        reference_id,
        created_at,
        expires_at,
        request_hash
    )
    SELECT
        gen_random_uuid(),
        gen_random_uuid(),
        gen_random_uuid(),
        NOW() - INTERVAL '4 days',
        NOW() - INTERVAL '3 days',
        -- RANDOM() generates a random decimal between 0 and 1
        -- ::text casts to a string
        -- MD5() gets the MD5 hash 0-9a-f, always lowercase in postgres
        MD5(RANDOM()::text)
    FROM generate_series(1, 300);

    -- add valid idempotency keys   ---------------------------------------------------------------------
    INSERT INTO public.idempotency_keys VALUES 
        (
            '019cbb9b-b299-7b00-bfcb-48dc05e218ae', 
            '019cbb9b-b284-73dc-9d2b-42d5ef94a0da', 
            '019cbb9b-b28c-7c35-9dc0-8f3c553ca432', 
            NOW(),
            NOW() + INTERVAL '2 days',
            '03625e1182b3954e91894bcb2609024067fa2a91bb51c5297393984a61a9f0b5'
        ),
        (
            '019cbb9b-df6f-7b4e-94e7-b1ca4479be3f', 
            '019cbb9b-df64-7084-ac7c-afd401688e47', 
            '019cbb9b-df69-755b-802f-ca1bed7fc4c9', 
            NOW(),
            NOW() + INTERVAL '2 days',
            'b9c1cadf995e713c65f79d1e8fe15842b195c4877564292f193cd659f5369a15'
        ),
        (
            '019cbb9c-3aeb-7340-9f6e-986d53058be6', 
            '019cbb9c-3ae0-734c-953d-debc4582beb9', 
            '019cbb9c-3ae6-7645-b54c-0b211b27cff7', 
            NOW(),
            NOW() + INTERVAL '2 days',
            'a1d2e626b63ef587d1b620b79a750402f0822833d7cf26712c7d75e6beae7b02'
        );
    INSERT INTO idempotency_keys (
        id,
        i_key,
        reference_id,
        created_at,
        expires_at,
        request_hash
    )
    SELECT
        gen_random_uuid(),
        gen_random_uuid(),
        gen_random_uuid(),
        NOW() - INTERVAL '4 days',
        NOW() + INTERVAL '24 hours',
        MD5(RANDOM()::text)
    FROM generate_series(1, 300);
-- ---------------------------------------------------------------------------------------------------------
    -- There should 3003 rows   
    -- add expired short urls ----------------------------------------------------------------------
    INSERT INTO short_urls (
        id,
        destination_url,
        slug,
        user_id,
        created_at,
        expires_at
    )
    SELECT
        gen_random_uuid(),
        SUBSTRING(MD5(RANDOM()::text) FROM 1 FOR 32) || '.example.invalid',
        SUBSTRING(MD5(RANDOM()::text) FROM 1 FOR 20),
        CASE
            WHEN RANDOM() < 0.5 THEN NULL
            ELSE (SELECT id FROM shurl_users WHERE i.n = i.n ORDER BY RANDOM() LIMIT 1) -- i.n = i.n forces postgres to reevaluate everytime
        END,
        NOW() - INTERVAL '4 days',
        NOW() - INTERVAL '8 days'
    FROM generate_series(1, 1000) AS i(n) -- i is the alias name of the result from gen_series and n is the name of the column of the result
    ON CONFLICT (slug) DO NOTHING;

    -- add valid short urls   ----------------------------------------------------------------------
    INSERT INTO public.short_urls VALUES 
        (
            '019cbb9b-b28c-7c35-9dc0-8f3c553ca432', 
            'https://google.com', 
            'tiLd', 
            NOW(),
            '019cb76d-23a3-7d94-9187-a702cbe03b3f', 
            NOW() + INTERVAL '7 days'
        ),
        (
            '019cbb9b-df69-755b-802f-ca1bed7fc4c9', 
            'https://gmail.com', 
            'L5y2', 
            NOW(),
            '019cb76d-23a3-7d94-9187-a702cbe03b3f', 
            NOW() + INTERVAL '7 days'
        ),
        (
            '019cbb9c-3ae6-7645-b54c-0b211b27cff7', 
            'https://mail.google.com', 
            'hBH2l', 
            NOW(),
            '019cb76d-23a3-7d94-9187-a702cbe03b3f', 
            NOW() + INTERVAL '7 days'
        );

    INSERT INTO short_urls (
        id,
        destination_url,
        slug,
        user_id,
        created_at,
        expires_at
    )
    SELECT
        gen_random_uuid(),
        SUBSTRING(MD5(RANDOM()::text) FROM 1 FOR 32) || '.example.invalid',
        SUBSTRING(MD5(RANDOM()::text) FROM 1 FOR 20),
        CASE
            WHEN RANDOM() < 0.5 THEN NULL
            ELSE (SELECT id FROM shurl_users WHERE i.n = i.n ORDER BY RANDOM() LIMIT 1)
        END,
        NOW() - INTERVAL '4 days',
        NOW() + INTERVAL '7 days'
    FROM generate_series(1, 2000) AS i(n)
    ON CONFLICT (slug) DO NOTHING;
-- ---------------------------------------------------------------------------------------------------------
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'Error happened %, rolling back...', SQLERRM;
    -- automatically aborts
END $$;
-- automatically commits at the end