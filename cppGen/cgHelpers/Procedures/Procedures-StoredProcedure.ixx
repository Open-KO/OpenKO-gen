	export class StoredProcedure
	{
	protected:
		StoredProcedure(std::shared_ptr<nanodbc::connection> conn)
			: _conn(conn), _stmt(*conn.get())
		{
			_flushed = false;
		}

		/// \brief Flushes any output variables or return values on destruction
		// This must be called in the destructor of a stored procedure with any output & return values.
		void flush_on_destruct()
		{
			try
			{
				flush();
			}
			// We should not throw exceptions from within a destructor.
			// We no longer care about the state of this statement anyway.
			catch (const nanodbc::database_error&)
			{
			}
		}

		/// \brief Executes the currently prepared statement
		/// \throws nanodbc::database_error
		/// \returns a result set, if applicable
		std::weak_ptr<nanodbc::result> execute() noexcept(false)
		{
			_flushed = false;
			_result = std::make_shared<nanodbc::result>(_stmt.execute());
			return _result;
		}

	public:
		/// \brief Flushes any output variables or return values by reading any and all result sets
        void flush()
        {
            if (_flushed
                || _result == nullptr)
                return;

            try
            {
                do
                {
                    skip_rows_in_result_set();
                }
                while (_result->next_result());
            }
            catch (const nanodbc::database_error& ex)
            {
                // This will trigger normally if no result sets are available,
                // which is typical behaviour for most stored procedures.
                if (ex.state() != SqlState_InvalidCursorState)
                    throw;
            }

            _flushed = true;
        }

	protected:
	    void skip_rows_in_result_set()
        {
            try
            {
                while (_result->next())
                {
                }
            }
            catch (const nanodbc::database_error& ex)
            {
                // This will trigger normally if no result sets are available,
                // which is typical behaviour for most stored procedures.
                if (ex.state() != SqlState_InvalidCursorState)
                    throw;
            }
        }

		std::shared_ptr<nanodbc::connection> _conn;
		nanodbc::statement _stmt;
		std::shared_ptr<nanodbc::result> _result;
		bool _flushed;

		static const nanodbc::string SqlState_InvalidCursorState;
	};

	const nanodbc::string StoredProcedure::SqlState_InvalidCursorState = NANODBC_TEXT("24000");
