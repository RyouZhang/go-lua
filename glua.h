#include <luajit.h>
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>

int gluaL_dostring(lua_State* _L, char* script);
void glua_getglobal(lua_State* _L, char* name);
void glua_setglobal(lua_State* _L, char* name);

void glua_pushlightuserdata(lua_State* _L, void* obj);
int glua_pcall(lua_State* _L, int args, int results);
lua_Number glua_tonumber(lua_State* _L, int index);
int glua_yield(lua_State *_L, int nresults);
const char* glua_tostring(lua_State* _L, int index);
void glua_pop(lua_State* _L, int num);
lua_State *glua_tothread(lua_State* _L, int index);
int glua_istable(lua_State* _L, int index);
void* glua_touserdata(lua_State* _L, int index);

//for go extra
void register_go_method(lua_State* _L);