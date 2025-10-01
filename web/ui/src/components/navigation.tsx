import { useState, useRef, useEffect } from "react";
import { ImageIcon, LayoutDashboardIcon, SearchIcon, TableIcon, ServerIcon, GlobeIcon, MenuIcon, XIcon } from "lucide-react";
import { Form, NavLink, useSubmit } from "react-router-dom";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { ModeToggle } from "./mode-toggle";
import { Popover, PopoverContent, PopoverTrigger } from "./ui/popover";
import { Badge } from "./ui/badge";

const navs = [
  { name: `Dashboard`, icon: <LayoutDashboardIcon className="mr-2 h-4 w-4" />, to: `/` },
  { name: `Gallery`, icon: <ImageIcon className="mr-2 h-4 w-4" />, to: `/gallery` },
  { name: `Overview`, icon: <TableIcon className="mr-2 h-4 w-4" />, to: `/overview` },
  { name: `IPs`, icon: <ServerIcon className="mr-2 h-4 w-4" />, to: `/ips` },
  { name: `Domains`, icon: <GlobeIcon className="mr-2 h-4 w-4" />, to: `/domains` },
];

const searchOperators = [
  { key: 'title', description: 'search by title' },
  { key: 'body', description: 'search by html body' },
  { key: 'tech', description: 'search by technology' },
  { key: 'header', description: 'search by header' },
  { key: 'p', description: 'search by perception hash' },
];

const Navigation = () => {
  const [searchValue, setSearchValue] = useState("");
  const [usedOperators, setUsedOperators] = useState<string[]>([]);
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const submit = useSubmit();

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchValue(e.target.value);
    const operators = e.target.value.match(/(\w+):/g) || [];
    setUsedOperators(operators.map(op => op.slice(0, -1)));
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsPopoverOpen(false);
    submit(event.currentTarget);
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (inputRef.current && !inputRef.current.contains(event.target as Node)) {
        setIsPopoverOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  return (
    <>
      <div className="relative">
        <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
          <div className="container flex h-auto min-h-16 items-center justify-between py-3">
            {/* Logo and Brand */}
            <NavLink to="/" className="flex items-center gap-4 px-2 mr-4 min-w-fit">
              <img src="/logo.png" alt="Defend Denmark ASM Logo" className="h-10 w-10 flex-shrink-0" />
              <span className="font-bold text-lg whitespace-nowrap" style={{ color: '#C60C30' }}>
                Defend Denmark ASM
              </span>
            </NavLink>

            {/* Desktop Navigation */}
            <div className="hidden lg:flex items-center gap-1">
              {navs.map(nav => {
                return <NavLink
                  key={nav.to}
                  to={nav.to}
                  className={({ isActive }) =>
                    isActive
                      ? "text-foreground transition-colors hover:text-foreground"
                      : "text-muted-foreground transition-colors hover:text-foreground"
                  }
                >
                  <Button variant="ghost" size="default">
                    {nav.icon} {nav.name}
                  </Button>
                </NavLink>;
              })}
            </div>

            {/* Mobile Menu Button */}
            <div className="lg:hidden flex items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                aria-label="Toggle menu"
              >
                {isMobileMenuOpen ? <XIcon className="h-5 w-5" /> : <MenuIcon className="h-5 w-5" />}
              </Button>
            </div>

            {/* Search and Mode Toggle - Desktop */}
            <div className="hidden lg:flex items-center gap-4 ml-auto">
              <Form method="post" action="/search" onSubmit={handleSubmit} className="flex-1 sm:flex-initial">
                <Popover open={isPopoverOpen}>
                  <PopoverTrigger asChild>
                    <div className="relative">
                      <SearchIcon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                      <Input
                        ref={inputRef}
                        name="query"
                        type="search"
                        placeholder="Search..."
                        className="pl-8 w-[300px]"
                        defaultValue={searchValue}
                        onChange={handleInputChange}
                        onFocus={() => {
                          setIsPopoverOpen(true);
                          setTimeout(() => {
                            if (inputRef.current) {
                              inputRef.current.focus();
                            }
                          }, 0);
                        }}
                      />
                    </div>
                  </PopoverTrigger>
                  <PopoverContent className="w-[300px] p-0" align="start">
                    <div className="grid gap-4 p-4">
                      <div className="space-y-2">
                        <h4 className="font-medium leading-none">Search Operators</h4>
                        <p className="text-sm text-muted-foreground">
                          Use these operators to refine your search.
                        </p>
                      </div>
                      <div className="grid gap-2">
                        {searchOperators.length === usedOperators.length && <div className="text-sm">No operators left.</div>}
                        {searchOperators.map((operator) => (
                          !usedOperators.includes(operator.key) && (
                            <div key={operator.key} className="flex items-center">
                              <Badge variant="secondary" className="mr-2">
                                {operator.key}:
                              </Badge>
                              <span className="text-sm">{operator.description}</span>
                            </div>
                          )
                        ))}
                      </div>
                    </div>
                  </PopoverContent>
                </Popover>
              </Form>
              <ModeToggle />
            </div>
          </div>
        </nav>

        {/* Mobile Menu Overlay */}
        {isMobileMenuOpen && (
          <div className="lg:hidden fixed inset-0 top-[73px] z-50 grid h-[calc(100vh-73px)] w-full grid-cols-1">
            <div className="bg-background border-r shadow-lg overflow-y-auto">
              <div className="p-6 space-y-3">
                {navs.map(nav => (
                  <NavLink
                    key={nav.to}
                    to={nav.to}
                    onClick={() => setIsMobileMenuOpen(false)}
                    className={({ isActive }) =>
                      `flex items-center px-4 py-3 rounded-lg transition-colors text-lg ${
                        isActive
                          ? "bg-accent text-foreground font-medium"
                          : "text-muted-foreground hover:bg-accent hover:text-foreground"
                      }`
                    }
                  >
                    {nav.icon} {nav.name}
                  </NavLink>
                ))}
                
                {/* Mobile Search */}
                <div className="pt-6">
                  <Form method="post" action="/search" onSubmit={(e) => { handleSubmit(e); setIsMobileMenuOpen(false); }}>
                    <div className="relative">
                      <SearchIcon className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                      <Input
                        name="query"
                        type="search"
                        placeholder="Search..."
                        className="pl-10 h-12 text-base"
                        defaultValue={searchValue}
                        onChange={handleInputChange}
                      />
                    </div>
                  </Form>
                </div>
                
                {/* Mobile Mode Toggle */}
                <div className="pt-4 flex justify-start">
                  <ModeToggle />
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
};

export default Navigation;