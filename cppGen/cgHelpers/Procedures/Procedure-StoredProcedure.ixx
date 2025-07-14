module;

#include <nanodbc/nanodbc.h>

export module Procedure:StoredProcedure;

namespace procedure {
    export class StoredProcedure
    {
    public:
        StoredProcedure(nanodbc::connection& conn)
            : _conn(conn), _stmt(conn)
        {
            _returnValue = 0;
            _finalized = false;
        }

        int returnValue()
        {
            finalize();
            return _returnValue;
        }

        void finalize()
        {
            if (_finalized
                || _result == nullptr)
                return;

            while (_result->next()
                || _result->next_result())
            {
            }

            _finalized = true;
        }

        ~StoredProcedure()
        {
            try
            {
                finalize();
            }
            catch (const nanodbc::database_error&)
            {
            }
        }

    protected:
        nanodbc::connection& _conn;
        nanodbc::statement _stmt;
        std::unique_ptr<nanodbc::result> _result;
        int _returnValue;
        bool _finalized;
    };
}
