#pragma once
#include "default_value.h"

namespace cgHelpers {
    template<typename T>
        class Nullable {
    public:
        void Set(T val) {
            Value = val;
            IsSet = true;
        }

        T Get() const {
            return Value;
        }

        bool IsValid() const {
            return IsSet;
        }

        void Reset() {
            Value = default_value<T>::value;
            IsSet = false;
        }

    private:
        T Value = default_value<T>::value;
        bool IsSet = false;
    };
}
