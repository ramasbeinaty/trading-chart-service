package db

func GetMigrationScripts() []MigrationScript {
	migrationScripts := []MigrationScript{
		{
			key: "initial",
			up: `
				SET TIMEZONE='UTC';

				CREATE OR REPLACE FUNCTION trigger_set_update_at()
				RETURNS TRIGGER AS $$
				BEGIN
					NEW.updated_at = NOW();
					RETURN NEW;
				END;
				$$ LANGUAGE plpgsql;
				
				CREATE TABLE IF NOT EXISTS candlestick (
					id SERIAL PRIMARY KEY,
					symbol VARCHAR(20) NOT NULL,
					open_price NUMERIC NOT NULL,
					high_price NUMERIC NOT NULL,
					low_price NUMERIC NOT NULL,
					close_price NUMERIC NOT NULL,
					trade_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
					created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),  
					updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
					UNIQUE (symbol, trade_timestamp)
				);
		`,
			down: `
				DROP FUNCTION trigger_set_update_at();
				DROP TABLE IF EXISTS candlestick;
		`,
		},
	}

	return migrationScripts
}
