module;

#include <nanodbc/nanodbc.h>
#include <memory>

export module Procedures:StoredProcedure;

namespace procedures {
    export class StoredProcedure
   {
   protected:
       StoredProcedure(nanodbc::connection& conn)
           : _conn(conn), _stmt(conn)
       {
           _flushed = false;
       }

       ~StoredProcedure()
       {
           try
           {
               flush();
           }
           catch (const nanodbc::database_error&)
           {
           }
       }

       std::weak_ptr<nanodbc::result> execute()
       {
           _flushed = false;
           _result = std::make_shared<nanodbc::result>(_stmt.execute());
           return _result;
       }

   public:
       void flush()
       {
           if (_flushed
               || _result == nullptr)
               return;

           while (_result->next()
               || _result->next_result())
           {
           }

           _flushed = true;
       }

   protected:
       nanodbc::connection& _conn;
       nanodbc::statement _stmt;
       std::shared_ptr<nanodbc::result> _result;
       bool _flushed;
   };
}
